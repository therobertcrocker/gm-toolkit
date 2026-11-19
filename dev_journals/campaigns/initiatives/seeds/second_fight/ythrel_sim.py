"""Ythrel, The Soft Collapse: strategy simulator.

Reuses the party and engine from gm-toolkit's hawk_finale_sim.py (same PC numbers,
same scoring), and swaps in Ythrel, her mental kit, Just Rest (calm), the Urgency
Vulnerability and the party's Name It action.

  python3 ythrel_sim.py --compare            # every Ythrel strategy side by side
  python3 ythrel_sim.py --strategy flowchart --log --seed 3
  python3 ythrel_sim.py --compare --set vuln_s=2 --set vuln_c=4
"""
import argparse, ast, sys, os, random, statistics as st
from collections import Counter
sys.path.insert(0, os.environ.get("SIM_DIR", "/home/claude/therobertcrocker/gm-toolkit/cmd/scripts"))
import hawk_finale_sim as H

d, degree = H.d, H.degree

DEFAULTS = dict(H.DEFAULTS)
DEFAULTS.update(H.RECOMMENDED)
DEFAULTS.update({
    "level": 14,
    "strategy": "sigh_brute",
    "ythrel_hp": 290, "ythrel_ac": 39,
    "dc": 37,            # her spell / breath / Unveil DC
    "dc_low": 35,        # Just Rest, Steal Knowledge
    "som": 1,            # Strength of Mind: +1 vs mental (+2 vs fear)
    "extra_will": 0,     # any further save bonus vs her (the unknown second buff)
    "just_rest": True,   # Diplomatic Solution on initiative
    "calm_rounds": 0,    # 0 = calm lasts until woken; N = wears off after N rounds
    "dom_rank_incap": True,  # rank-6 Dominate: PCs' results one step better
    "vuln_s": 3, "vuln_c": 6,   # d6s on a PC success / crit success
    "name_it": 4,        # d6s for Name It
    "stay": True, "stay_at": 75,
    "lull_touch": False,
    # positioning bands, relative to Ythrel: 0 Engaged (her reach), 1 Reachable (one Stride), 2 Far
    "bands": False,
    "party_pos": "smart",  # bunch | spread | smart (ranged PCs back off to Far after the first Sigh)
    "sigh_bands": 1,       # the Sigh catches everyone in bands <= this (1 = Engaged + Reachable)
    "fold": False,         # end of each round, the Lull pulls Far PCs to Reachable
    "start_band": 1,
    "y_move": "none",
    "hunt": False,
    "crit_is_turning_point": False,  # scoring: a 12d6 flat refusal counts as the fight's turning point
    "surround_dice": 0,    # d6s she takes at the end of her turn if 2+ PCs are Engaged with her
    "surround_with": "tavish",  # once found: who joins Kas to surround her (tavish | both)
    "escape": False,       # she pours away when she'd end her turn surrounded         # with pour: she leaves Kas to chase a backliner who hurt her (only when the party isn't struggling)      # none | pour (she moves: 1 action re-centres on a PC) | pull (Come Here, Dear) | both
    "atk_bonus": 0, "dmg_bonus": 0,   # added to every Strike's attack / damage
    "sigh_dice": 15, "sigh_recharge": 4,
    "ythrel_fort": 28,
    "stay_temp": 20,
    "self_trigger": 0,    # d6s she takes at end of turn if she Strikes a PC who hasn't harmed her since her last turn
    "target_rule": "any", # any | prefer (harmers first) | strict (only harmers; else Bless Your Heart)  # first claw/tail hit each turn: Will save or slowed 1 (crit fail slowed 2)
    "reach_p": 0.5,      # chance Tavish/Arabesque are inside her reach on a given turn
    "start_frac": 1.0,   # party HP fraction at the start (marathon wear)
    "disc_curve": None,  # {round: chance} the party finds Name It; default = hint ladder
})

# The build on the GM flowchart (in-benchmark, level 16). Running with no --set reproduces the chart.
CHART_BUILD = dict(strategy="sigh_brute", just_rest=False, stay=False, bands=True, party_pos="smart", y_move="pour",
                   fold=False, ythrel_hp=299, ythrel_ac=37, dc=34, dc_low=32, sigh_dice=17, dmg_bonus=3, atk_bonus=2,
                   lull_touch=True, target_rule="prefer", name_it=0, vuln_s=0, vuln_c=4, self_trigger=4,
                   surround_dice=4, surround_with="both", escape=False, disc_curve={1: 0.6, 2: 1.0})
DEFAULTS.update(CHART_BUILD)

STRATEGIES = {
    "flowchart":   "the published chart: Whispers when 2+ calmed > Sigh+Steal > Dominate+Claw > Unveil+Jaws > Frenzy+Steal",
    "brute":       "Strikes only (Frenzy + Jaws); mental only when nobody is in reach",
    "lean_mental": "Sigh when ready > Unveil each PC once (+Jaws) > Frenzy + Jaws; no Steal, no Whispers, no Dominate",
    "dominator":   "Unveil a PC, then Dominate them using the reveal; Sigh when ready; else Frenzy + Jaws",
    "sigh_brute":  "Sigh whenever ready (+Jaws/Claw), otherwise Frenzy + Jaws",
    "sigh_dom":    "Sigh when ready (+Jaws) > Dominate the hardest hitter if they're winning (+Claw) > Frenzy + Jaws",
}

R = H.R
CFG = H.CFG

def setcfg(over):
    CFG.clear(); CFG.update(DEFAULTS); CFG.update(over)

# --------------------------------------------------------------- saves vs her
def psave(pc, dc, ctx, fear=False, label="", reveal_ok=True):
    """A PC's Will save against one of Ythrel's mental effects. Fires the vulnerability."""
    bonus = CFG["som"] * (2 if fear else 1) + CFG["extra_will"]
    stup = pc.stupefied[0] if pc.stupefied[1] else 0
    mod = pc.sv["will"] + bonus - pc.frightened - pc.sickened - stup
    r = R.randint(1, 20)
    if reveal_ok and pc.secret and ctx["ythrel"].reaction:
        r = min(r, R.randint(1, 20)); pc.secret = False; ctx["ythrel"].reaction = False
        log(f"     Ythrel reveals {pc.name}'s secret: roll twice, take the lower")
    g = H.adj("pc_save", degree(r, r + mod, dc))
    if fear and getattr(pc, "bravery", False) and g == 2: g = 3
    STATS["psaves"][g] += 1
    if g >= 2: vuln(pc, g, ctx, label)
    return g

def vuln(pc, g, ctx, label):
    if g == 2 and CFG["vuln_s"] == 0: return
    if pc.name in ctx["vuln_used"]: return
    ctx["vuln_used"].add(pc.name)
    y = ctx["ythrel"]
    dmg = y.take(d(CFG["vuln_c"] if g == 3 else CFG["vuln_s"], 6), src="Vuln")
    STATS["vuln_hits"] += 1; ctx["flaked"] = True
    if CFG.get("crit_is_turning_point") and g == 3: H.TRACK["holy_hit"] = True; ctx["crit_tp"] = True
    log(f"     {pc.name} refuses ({label}): Conflict Vulnerability, Ythrel takes {dmg} (gilt flakes)")

def log(s): H.log(s)

# ------------------------------------------------------------- Ythrel's moves
def up(party): return [p for p in party if p.up]
def awake(party): return [p for p in party if p.up and not p.calmed and not p.controlled]
def uncalmed(party): return [p for p in party if p.up and not p.calmed]

def reachable(party):
    if CFG["bands"]: r = [p for p in uncalmed(party) if p.band == 0]
    else: r = [p for p in uncalmed(party) if p.name == "Kas" or R.random() < CFG["reach_p"]]
    rule = CFG["target_rule"]
    harmers = [p for p in r if p.harmed]
    if rule == "prefer" and harmers: return harmers
    if rule == "strict": return harmers
    return r

def momentum(t, g):
    if g == 3: CTX["sigh_cd"] = 0
    if CFG["lull_touch"] and not CTX.get("touched") and t.up:
        CTX["touched"] = True
        sg = psave(t, CFG["dc"], CTX, label="There, There")
        if sg <= 1: t.slowed = max(t.slowed, 2 if sg == 0 else 1); log(f"     There, There: {t.name} slowed {t.slowed}")

def strike(y, t, which, ctx, map_=0):
    stats = {"claw": (32, lambda: d(3, 10) + 15, 4), "tail": (30, lambda: d(3, 10) + 15, 5),
             "jaws": (32, lambda: d(3, 12) + 15, 5)}
    mod, dmg0, _ = stats[which]
    mod += CFG["atk_bonus"]; dmg = lambda: dmg0() + CFG["dmg_bonus"]
    if t is not None and t.up and not t.harmed: ctx["unprovoked"] = True
    return H.boss_strike(y, t, mod - map_, dmg, which.capitalize(), ctx, momentum)

def frenzy(y, t, ctx):
    log(f"  Ythrel: TUCK IN (claw, claw, tail) on {t.name}")
    strike(y, t, "tail", ctx, 0)
    if t.up: strike(y, t, "claw", ctx, 4)
    if t.up: strike(y, t, "claw", ctx, 8)
    return 3   # attacks made, for MAP

def sigh_pool(party):
    pool = [p for p in uncalmed(party) if p.band <= CFG["sigh_bands"]] if CFG["bands"] else uncalmed(party)
    if CFG["wail_mercy"]: pool = [p for p in pool if p.hp >= 0.4 * p.maxhp]
    return pool

def sigh(y, party, ctx):
    pool = sigh_pool(party) or uncalmed(party)
    if CFG["bands"]: tg = pool
    else: tg = R.sample(pool, 3 if len(pool) == 3 and R.random() < 0.5 else min(2, len(pool)))
    STATS["sigh_n"] += 1; STATS["sigh_hit"] += len(tg); ctx["sigh_seen"] = True
    dmg = d(CFG["sigh_dice"], 6)
    log(f"  Ythrel: THE LONG SIGH ({dmg} mental) on {', '.join(t.name for t in tg)}")
    for t in tg:
        g = psave(t, CFG["dc"], ctx, label="Long Sigh")
        m = [2, 1, 0.5, 0][g]; t.take(int(dmg * m), "Long Sigh", crit=(g == 0))
        if g <= 1 and t.up: t.stupefied = (2 if g == 0 else 1, 10)
        log(f"     {t.name}: {['CRIT FAIL', 'fail', 'success', 'crit success'][g]} -> {int(dmg * m)} ({max(t.hp, 0)}/{t.maxhp})")
    ctx["sigh_cd"] = R.randint(1, CFG["sigh_recharge"])

def steal(y, party, ctx):
    pool = awake(party)
    if not pool: return
    t = R.choice(pool); psave(t, CFG["dc_low"], ctx, label="You Don't Need That")
    log(f"  Ythrel: YOU DON'T NEED THAT on {t.name}")

def unveil(y, t, ctx):
    log(f"  Ythrel: BLESS YOUR HEART on {t.name}")
    t.cut = True
    g = psave(t, CFG["dc"], ctx, fear=True, label="Bless Your Heart", reveal_ok=False)
    if g <= 1: t.frightened = max(t.frightened, 1); t.sticky = 10; t.secret = True

def dominate(y, t, ctx):
    ctx["dom_uses"] -= 1
    log(f"  Ythrel: SIT DOWN, SWEETHEART on {t.name}")
    g = psave(t, CFG["dc"], ctx, label="Sit Down")
    if CFG["dom_rank_incap"] and g < 3:
        g2 = g + 1
        if g2 >= 2 and g < 2: vuln(t, g2, ctx, "Sit Down (incapacitation)")
        g = g2
    if g == 2: t.stunned += 1; log(f"     {t.name} stunned 1")
    elif g <= 1: t.controlled = 2 if g == 0 else 1; STATS["controlled"] += 1; log(f"     {t.name} is CONTROLLED")

def whispers(y, party, ctx):
    ctx["whispers"] = True
    log("  Ythrel: THOUGHT WHISPERS on everyone")
    for t in up(party): psave(t, CFG["dc"], ctx, label="Thought Whispers")

def stay(y, party, ctx):
    ctx["stayed"] = True
    log("  Ythrel: STAY")
    for t in up(party):
        g = psave(t, CFG["dc"], ctx, label="Stay")
        if g == 1: t.slowed = max(t.slowed, 2); t.temp = max(t.temp, CFG["stay_temp"]); log(f"     {t.name} slowed 2, +{CFG['stay_temp']} temp")
        if g == 0: t.lost_turn = 1; log(f"     {t.name} falls asleep")

def pour(y, party, t, ctx):
    """1 action: she pours across the floor to t. t becomes Engaged; PCs who were in
    t's band end up Reachable; everyone else ends up Far."""
    b = t.band
    for p in party:
        if p is t: continue
        p.band = 1 if p.band == b else 2
    t.band = 0
    STATS["pours"] += 1
    log(f"  Ythrel POURS across the floor to {t.name}")

def pull(y, t, ctx):
    """1 action, Come Here, Dear: Will save or slide one band closer (crit fail: Engaged)."""
    STATS["pulls"] += 1
    g = psave(t, CFG["dc"], ctx, label="Come Here, Dear")
    if g <= 1 and t.band > 0:
        t.band = 0 if g == 0 else t.band - 1
        log(f"  Ythrel: COME HERE, DEAR. {t.name} slides to {['Engaged', 'Reachable', 'Far'][t.band]}")
    else:
        log(f"  Ythrel: COME HERE, DEAR. {t.name} holds their ground")

def mobile_turn(y, party, ctx):
    mv = CFG["y_move"]; a = 3
    unc = uncalmed(party)
    ctx["acts_left"] = 0
    reserve = CFG["escape"] and CFG["surround_dice"] and len([p for p in up(party) if p.band == 0]) >= 2
    def pool(): return sigh_pool(party)
    # 1. Sigh: bunch them first if she has to
    if ctx["sigh_cd"] == 0 and len([p for p in unc if p.hp >= 0.4 * p.maxhp]) >= 2:
        if len(pool()) < 2 and mv in ("pull", "both"):
            out = [p for p in unc if p.band > CFG["sigh_bands"] and p.hp >= 0.4 * p.maxhp]
            if out: pull(y, R.choice(out), ctx); a -= 1
        if len(pool()) < 2 and mv in ("pour", "both") and a == 3:
            groups = {}
            for p in unc:
                if p.hp >= 0.4 * p.maxhp: groups.setdefault(p.band, []).append(p)
            big = max(groups.values(), key=len) if groups else []
            if len(big) >= 2: pour(y, party, R.choice(big), ctx); a -= 1
        if len(pool()) >= 2 and a >= 2:
            sigh(y, party, ctx); a -= 2
            eng = [p for p in reachable(party)]
            if a and reserve: ctx["acts_left"] = a; return
            if a and eng: strike(y, pick_target(eng), "jaws", ctx)
            return
    # 2. Claws: get someone into reach (prefer whoever hurt her)
    eng = reachable(party)
    if CFG["hunt"] and mv in ("pour", "both") and a == 3 and not H.struggling(party):
        back = [p for p in unc if p.band > 0 and p.harmed]
        if back:
            pour(y, party, max(back, key=lambda p: p.hp), ctx); a -= 1; STATS["hunts"] += 1
            eng = reachable(party)
    if not eng:
        cands = [p for p in unc if p.harmed] or unc
        if cands and mv in ("pour", "both"):
            pour(y, party, pick_target(cands), ctx); a -= 1
        elif cands and mv == "pull":
            t = pick_target(cands); pull(y, t, ctx); a -= 1
            if t.band > 0: unveil(y, t, ctx); return
        eng = reachable(party)
    if eng and a >= 2:
        t = pick_target(eng); frenzy(y, t, ctx)
        if a == 3 and reserve: ctx["acts_left"] = 1
        elif a == 3 and t.up: strike(y, t, "jaws", ctx, 10)
    elif unc and a >= 2:
        unveil(y, R.choice(unc), ctx)

def pick_target(cands):
    w = {"Kas": 4, "Tavish": 3, "Arabesque": 3}
    return R.choices(cands, [w[c.name] for c in cands])[0] if cands else None

def ythrel_turn(y, party, ctx):
    ctx["unprovoked"] = False
    _ythrel_turn(y, party, ctx)
    if CFG["surround_dice"] and CFG["bands"] and y.hp > 0:
        eng = [p for p in up(party) if p.band == 0]
        if len(eng) >= 2 and CFG["escape"] and ctx.get("acts_left", 0) >= 1:
            pour(y, party, R.choice(eng), ctx); STATS["escapes"] += 1
            eng = [p for p in up(party) if p.band == 0]
        if len(eng) >= 2:
            dmg = y.take(d(CFG["surround_dice"], 6), src="Vuln surround")
            STATS["surrounds"] += 1; ctx["flaked"] = True
            if ctx["discovered"]: H.TRACK["holy_hit"] = True
            log(f"   Ythrel is SURROUNDED ({', '.join(p.name for p in eng)}): Conflict Vulnerability, she takes {dmg}")
    if CFG["self_trigger"] and ctx["unprovoked"] and y.hp > 0:
        dmg = y.take(d(CFG["self_trigger"], 6), src="Vuln self")
        STATS["self_hits"] += 1
        log(f"   Ythrel struck someone who never touched her: Conflict Vulnerability, she takes {dmg}")
    for p in party: p.harmed = False

def _ythrel_turn(y, party, ctx):
    y.reaction = True
    y.prone = False; y.clumsy = 0; y.enfeebled = 0; y.frightened = max(0, y.frightened - 1)
    if not up(party): return
    ctx["touched"] = False
    if CFG["stay"] and not ctx["stayed"] and y.hp <= CFG["stay_at"]:
        stay(y, party, ctx); return
    s = CFG["strategy"]
    calmed = [p for p in up(party) if p.calmed]
    reach = reachable(party)
    unc = uncalmed(party)
    sigh_ok = ctx["sigh_cd"] == 0 and (len(sigh_pool(party)) >= 2 if CFG["bands"] else len(unc) >= 2)
    if CFG["bands"]:
        STATS["turns"] += 1
        if len([p for p in up(party) if p.band <= CFG["sigh_bands"]]) >= 2: STATS["bunched"] += 1
    if s == "flowchart":
        if len(calmed) >= 2:
            if not ctx["whispers"]: whispers(y, party, ctx)
            else: log("  Ythrel waits. Time passes pleasantly.")
            return
        if sigh_ok: sigh(y, party, ctx); steal(y, party, ctx); return
        hard = ctx.get("hardest")
        if hard and hard.up and not hard.calmed and ctx["dom_uses"] and not H.struggling(party):
            dominate(y, hard, ctx)
            if reach: strike(y, pick_target(reach), "claw", ctx)
            return
        nocut = [p for p in unc if not p.cut]
        if nocut:
            t = R.choice(nocut); unveil(y, t, ctx)
            if reach: strike(y, pick_target(reach), "jaws", ctx)
            return
        if reach: frenzy(y, pick_target(reach), ctx)
        steal(y, party, ctx); return
    if s == "brute":
        if reach:
            t = pick_target(reach); frenzy(y, t, ctx)
            if t.up: strike(y, t, "jaws", ctx, 10)
            return
        if sigh_ok: sigh(y, party, ctx); return
        if unc: unveil(y, R.choice(unc), ctx)
        return
    if s == "sigh_brute" and CFG["bands"] and CFG["y_move"] != "none":
        return mobile_turn(y, party, ctx)
    if s == "sigh_brute":
        if sigh_ok:
            sigh(y, party, ctx)
            if reach: strike(y, pick_target(reach), "jaws", ctx)
            return
        if reach:
            t = pick_target(reach); frenzy(y, t, ctx)
            if t.up: strike(y, t, "jaws", ctx, 10)
        elif unc: unveil(y, R.choice(unc), ctx)
        return
    if s == "sigh_dom":
        if sigh_ok:
            sigh(y, party, ctx)
            if reach: strike(y, pick_target(reach), "jaws", ctx)
            return
        hard = ctx.get("hardest")
        if hard and hard.up and not hard.calmed and not hard.controlled and ctx["dom_uses"] and not H.struggling(party):
            dominate(y, hard, ctx)
            if reach: strike(y, pick_target(reach), "claw", ctx)
            return
        if reach:
            t = pick_target(reach); frenzy(y, t, ctx)
            if t.up: strike(y, t, "jaws", ctx, 10)
        elif unc: unveil(y, R.choice(unc), ctx)
        return
    if s == "lean_mental":
        if sigh_ok:
            sigh(y, party, ctx)
            if reach: strike(y, pick_target(reach), "jaws", ctx)
            return
        nocut = [p for p in unc if not p.cut]
        if nocut:
            unveil(y, R.choice(nocut), ctx)
            if reach: strike(y, pick_target(reach), "jaws", ctx)
            return
        if reach:
            t = pick_target(reach); frenzy(y, t, ctx)
            if t.up: strike(y, t, "jaws", ctx, 10)
        return
    if s == "dominator":
        if sigh_ok and R.random() < 0.5:
            sigh(y, party, ctx)
            if reach: strike(y, pick_target(reach), "jaws", ctx)
            return
        secret = [p for p in unc if p.secret and not p.controlled]
        if secret and ctx["dom_uses"]:
            t = max(secret, key=lambda p: p.name == "Kas"); dominate(y, t, ctx)
            if reach: strike(y, pick_target(reach), "claw", ctx)
            return
        nocut = [p for p in unc if not p.cut]
        if nocut and ctx["dom_uses"]:
            t = next((p for p in nocut if p.name == "Kas"), R.choice(nocut)); unveil(y, t, ctx)
            if reach: strike(y, pick_target(reach), "jaws", ctx)
            return
        if sigh_ok: sigh(y, party, ctx); return
        if reach:
            t = pick_target(reach); frenzy(y, t, ctx)
            if t.up: strike(y, t, "jaws", ctx, 10)
        return

# --------------------------------------------------------------- party turns
def name_it(pc, party, ctx, self_first=False):
    """1 action: 4d6 mental to Ythrel and clear one of her conditions."""
    if pc.name in ctx["named"] or not ctx["discovered"]: return False
    ctx["named"].add(pc.name)
    y = ctx["ythrel"]; dmg = y.take(d(CFG["name_it"], 6), src="Name It")
    H.TRACK["holy_hit"] = True; STATS["name_it"] += 1
    order = ([pc] if self_first else []) + ([p for p in party if p.controlled and p.up] + [p for p in party if p.calmed and p.up] +
             [p for p in party if p.sticky and p.up] + [p for p in party if p.stupefied[1] and p.up])
    cleared = ""
    if order:
        t = order[0]
        if t.controlled: t.controlled = 0; cleared = f", frees {t.name}"
        elif t.calmed: t.calmed = 0; cleared = f", wakes {t.name}"
        elif t.sticky: t.sticky = 0; t.frightened = 0; t.secret = False; cleared = f", steadies {t.name}"
        else: t.stupefied = (0, 0); cleared = f", clears {t.name}'s head"
    log(f"   {pc.name} NAMES IT: Ythrel takes {dmg}{cleared}")
    return True

def pc_turn(pc, party, ctx):
    y = ctx["ythrel"]
    if pc.controlled:
        log(f"   {pc.name} sits down, as told")
        pc.slowed = 0; pc.stunned = 0
        g = psave(pc, CFG["dc"], ctx, label="end of turn vs Sit Down")
        if g >= 2: pc.controlled = 0; log(f"   {pc.name} stands back up")
        return
    if pc.lost_turn:
        pc.lost_turn -= 1; log(f"   {pc.name} is asleep"); pc.slowed = 0; pc.stunned = 0; return
    if pc.calmed:
        if ctx["discovered"]:
            # wake yourself, then act with the rest of the turn
            pc.extra_cost = 1; name_it(pc, party, ctx, self_first=True)
        else:
            log(f"   {pc.name} is calm and does nothing hostile"); pc.slowed = 0; pc.stunned = 0
            if pc.name == "Arabesque" and not ctx["haste_cast"] and CFG["haste"]:
                ctx["haste_cast"] = True
                for p in party: p.hasted = p.up and p.name != "Arabesque"
                log("   Arabesque casts haste (not hostile)")
            return
    # A free action of urgency: once known, everyone Names It (costs an action)
    cost = getattr(pc, "extra_cost", 0); pc.extra_cost = 0
    if ctx["discovered"] and CFG["name_it"] and pc.name not in ctx["named"]:
        cost += 1; name_it(pc, party, ctx)
    pc.pre_cost = cost
    hp0 = y.hp
    if ctx["named"] and pc.name in ctx["named"]: pc.harmed = True
    if pc.name == "Kas": kas_turn(pc, y, ctx)
    elif pc.name == "Tavish": tav_turn(pc, party, y, ctx)
    else: ara_turn(pc, party, y, ctx)
    dealt = hp0 - y.hp
    if dealt > 0: pc.harmed = True
    if dealt > ctx["round_best"][1]: ctx["round_best"] = (pc, dealt)

def move(pc, to, a, why=""):
    """Spend actions to change band. Returns actions left."""
    steps = abs(pc.band - to)
    if not CFG["bands"] or steps == 0: return a
    if pc.name == "Kas" and pc.band == 2 and to == 0: steps = 1     # Sudden Charge
    steps = min(steps, a)
    pc.band += steps if to > pc.band else -steps
    STATS["move_actions"] += steps
    log(f"   {pc.name} moves to {['Engaged', 'Reachable', 'Far'][pc.band]}{why}")
    return a - steps

def want_band(pc, ctx):
    if pc.name == "Kas": return 0
    if CFG["surround_dice"] and ctx.get("discovered"):
        if pc.name == "Tavish" or CFG["surround_with"] == "both": return 0
    pol = CFG["party_pos"]
    if pol == "bunch": return 1
    if pol == "spread": return 2
    return 2 if ctx.get("sigh_seen") else 1

def acts(pc):
    a = H.actions_for(pc) - pc.pre_cost; pc.pre_cost = 0
    return max(0, a)

def kas_turn(kas, y, ctx):
    a = acts(kas); party = ctx["party"]
    a = move(kas, 0, a)
    hurt = [p for p in party if not p.dead and (p.hp <= 0 or p.hp < 0.5 * p.maxhp)]
    if kas.wave and a >= 2 and (len(hurt) >= 2 or any(p.hp <= 0 and not p.dead for p in party)):
        kas.wave = False; a -= 2
        for p in party:
            near = (p.band <= 1) if CFG["bands"] else (R.random() < 0.7)
            if not p.dead and (p is kas or near): p.heal(d(kas.lvl, 4))
        log("   Kas activates HEALING WAVE")
    elif kas.hp < CFG["heal_thresh"] * kas.maxhp and kas.potions and a:
        h = kas.potions.pop(0); kas.heal(h); a -= 1; log(f"   Kas drinks a potion (+{h})")
    worse = False
    map_ = 0
    if a >= 1 and y.reaction and ctx["calm_known"]: pass
    pen = 2 if y.reaction else 0   # Shh. on his first attack
    if pen: y.reaction = False
    if a >= 2 and y.hp > 0:
        H.kas_hit(kas, y, ctx, kas.atk - pen, 2, "Vicious Swing"); a -= 2; map_ = 2; pen = 0
    while a > 1 and y.hp > 0 and kas.hasted:
        H.kas_hit(kas, y, ctx, kas.atk - pen - min(10, 5 * map_), 0, "Strike"); a -= 1; map_ += 1; pen = 0
    if a >= 1 and y.hp > 0:
        H.kas_hit(kas, y, ctx, kas.atk - pen - min(10, 5 * map_), 1, "Brutal Finish", brutal=True)

def tav_turn(tav, party, y, ctx):
    a = acts(tav)
    tgt = next((p for p in party if p is not tav and not p.dead and
                (p.hp <= 0 or p.hp < CFG["heal_thresh"] * p.maxhp)), None)
    if tgt and tav.elixirs and a >= 2 + (abs(tav.band - tgt.band) if CFG["bands"] else 0):
        a = move(tav, tgt.band, a, f" to reach {tgt.name}")
        h = d(5, 6) + 25; tav.elixirs -= 1; tgt.heal(h); a -= 2
        log(f"   Tavish feeds {tgt.name} an elixir of life (+{h})")
    elif tav.hp < CFG["heal_thresh"] * tav.maxhp and tav.potions and a:
        h = tav.potions.pop(0); tav.heal(h); a -= 1; log(f"   Tavish drinks a potion (+{h})")
    if CFG["bands"] and a >= 2: a = move(tav, want_band(tav, ctx), a)
    n = 0
    while a > 0 and y.hp > 0:
        mod = tav.atk - (0 if n == 0 else 5 if n == 1 else 10)
        if n == 0 and y.reaction: mod -= 2; y.reaction = False
        relic = None
        if n == 0 and a >= 2:
            relic = "tidal" if tav.tidal else "flare"
        g, _ = H.strike(tav, y, mod)
        dmg = {3: d(3, 6) * 2 + 5, 2: d(3, 6) + 5, 1: 5, 0: 0}[g]
        if relic and g >= 2: dmg += (d(8, 10) if relic == "tidal" else d(7, 6)) * (2 if g == 3 else 1)
        if relic == "tidal": tav.tidal = False
        dealt = y.take(dmg, src="Tavish bombs")
        log(f"   Tavish {relic or 'bomb'} {['crit-misses', 'misses (splash)', 'hits', 'CRITS'][g]} for {dealt}")
        a -= 2 if relic else 1; n += 1

def ara_turn(ara, party, y, ctx):
    a = acts(ara)
    if CFG["bands"] and a >= 3: a = move(ara, want_band(ara, ctx), a)
    if CFG["haste"] and not ctx["haste_cast"] and a >= 2:
        ctx["haste_cast"] = True; a -= 2
        for p in party: p.hasted = p.up and p.name != "Arabesque"
        log("   Arabesque casts haste on Kas and Tavish")
    if a >= 2 and ara.slots and y.hp > 0:
        rank = ara.slots.pop(0); a -= 2
        sv, sr = ara.stupefied
        if sr and R.randint(1, 20) < 5 + sv:
            log("   Arabesque's spell is lost to stupefied"); return
        dmg = d(2 * rank, 6)
        g = y.save("fort", ara.dc - ara.frightened - (sv if sr else 0))
        dealt = y.take(int(dmg * [2, 1, 0.5, 0][g]), src="Arabesque spells")
        ara.temp = max(ara.temp, dealt // 2)
        log(f"   Arabesque: vampiric exsanguination {['crit fail', 'fail', 'success', 'crit success'][g]}, {dealt}")
    if ara.hp < 0.4 * ara.maxhp and a and party[1].potions:
        h = party[1].potions.pop(0); ara.heal(h); log(f"   Arabesque drinks a potion (+{h})")

# ------------------------------------------------------------------ the fight
STATS = None
CTX = None

def new_stats():
    s = H.new_stats(); s.update(psaves=Counter(), vuln_hits=0, name_it=0, controlled=0, calm=Counter(),
                                calm_lock=0, self_hits=0, turns=0, bunched=0, sigh_n=0, sigh_hit=0, move_actions=0, pours=0, pulls=0, hunts=0, surrounds=0, escapes=0)
    return s

def end_of_turn(p):
    H.end_of_turn(p)
    if p.sticky:
        p.sticky -= 1; p.frightened = max(p.frightened, 1)
        if not p.sticky: p.secret = False

def fight():
    global CTX
    party = H.make_party()
    for p in party:
        p.hp = int(p.maxhp * CFG["start_frac"])
        p.calmed = 0; p.controlled = 0; p.cut = False; p.secret = False; p.sticky = 0
        p.extra_cost = 0; p.pre_cost = 0; p.harmed = False; p.band = CFG["start_band"]
    y = H.Boss("Ythrel", CFG["ythrel_hp"], CFG["ythrel_ac"], CFG["ythrel_fort"], 26, 30)
    ctx = {"party": party, "phase": 3, "ythrel": y, "sigh_cd": 0, "dom_uses": 2, "whispers": False,
           "stayed": False, "discovered": False, "discovered_round": None, "haste_cast": False,
           "vuln_used": set(), "named": set(), "flaked": False, "hardest": None,
           "round_best": (None, 0), "calm_known": False, "kas_move": 0}
    CTX = ctx
    if CFG["just_rest"]:
        log("Initiative: JUST REST (calm, DC %d)" % CFG["dc_low"])
        for p in party:
            g = psave(p, CFG["dc_low"], ctx, label="Just Rest")
            if g <= 1: p.calmed = 1; STATS["calm"][p.name] += 1; log(f"   {p.name} is calmed")
        if all(p.calmed for p in party): STATS["calm_lock"] += 1
    ctx["vuln_used"] = set()
    order = party + [y]; R.shuffle(order)
    ctx["order"] = order
    return _rounds(party, y, ctx, order, 1)

def _rounds(party, y, ctx, order, start):
    prev = None
    for rnd in range(start, 16):
        if prev is not None and prev - max(y.hp, 0) >= 0.35 * y.maxhp: H.TRACK["comeback"] = True
        prev = max(y.hp, 0)
        ctx["round"] = rnd; ctx["named"] = set(); ctx["round_best"] = (None, 0)
        if rnd > 1: ctx["vuln_used"] = set()
        log(f"-- Round {rnd}: Ythrel {max(y.hp, 0)}/{y.maxhp} | " +
            " | ".join(f"{p.name} {'DEAD' if p.dead else max(p.hp, 0)}" +
                       (" calm" if p.calmed else "") + (" CTRL" if p.controlled else "") for p in party))
        for c in order:
            if y.hp <= 0: return finish(party, ctx, rnd)
            if not any(p.up for p in party): return finish(party, ctx, -rnd)
            if c is y:
                ythrel_turn(y, party, ctx)
                if ctx["sigh_cd"]: ctx["sigh_cd"] -= 1
            else:
                if c.dead: continue
                if c.hp <= 0: H.dying_check(c); continue
                log(f"  {c.name}:"); pc_turn(c, party, ctx); end_of_turn(c)
        H.round_end(party, max(y.hp, 0) / y.maxhp)
        if y.hp <= 0: return finish(party, ctx, rnd)
        if not any(p.up for p in party): return finish(party, ctx, -rnd)
        ctx["hardest"] = ctx["round_best"][0]
        if CFG["bands"] and CFG["fold"]:
            for p in party:
                if p.band == 2 and p.up: p.band = 1; log(f"   The Lull folds: {p.name} is Reachable again")
        if CFG["calm_rounds"] and rnd >= CFG["calm_rounds"]:
            for p in party: p.calmed = 0
        if not ctx["discovered"]:
            curve = CFG["disc_curve"] or {1: 0.0, 2: 0.25, 3: 0.5, 4: 0.6, 5: 1.0}
            p = curve.get(rnd, 1.0)
            if rnd < 3 and not ctx["flaked"]: p = 0
            if R.random() < p:
                ctx["discovered"] = True; ctx["discovered_round"] = rnd
                log(f"  ** The party works out the weakness (end of round {rnd})")
        if WINCHANCE_N and not IN_ROLLOUT:
            H.TRACK["wc"].append(win_chance(party, y, ctx))
    return finish(party, ctx, 99)

# --------------------------------------------------------------- win chance
WINCHANCE_N = 0       # rollouts per round end; 0 = off. Set with --winchance N.
IN_ROLLOUT = False

def win_chance(party, y, ctx):
    """Copy the fight at this round end and play it out WINCHANCE_N times."""
    global STATS, CTX, IN_ROLLOUT
    import copy
    saved = (STATS, H.STATS, H.TRACK, CTX, H.LOGGING)
    IN_ROLLOUT = True; H.LOGGING = False; wins = 0
    try:
        for _ in range(WINCHANCE_N):
            p2, y2, c2, st2, tr2 = copy.deepcopy((party, y, ctx, saved[0], saved[2]))
            STATS = H.STATS = st2; H.TRACK = tr2; CTX = c2
            res = _rounds(p2, y2, c2, c2["order"], c2["round"] + 1)
            wins += res["result"] == "win"
    finally:
        STATS, H.STATS, H.TRACK, CTX, H.LOGGING = saved; IN_ROLLOUT = False
    return wins / WINCHANCE_N

def finish(party, ctx, r):
    if H.TRACK is not None and (not H.TRACK["diffs"] or ctx.get("last_end") != r):
        ctx["last_end"] = r; H.round_end(party, max(ctx["ythrel"].hp, 0) / ctx["ythrel"].maxhp)
    res = {"p1": r, "p2": None, "deaths": [p.name for p in party if p.dead],
           "discovered_round": ctx["discovered_round"]}
    res["result"] = "loss" if r < 0 else ("win" if r != 99 else "stalemate")
    return res

def run_many(n):
    global STATS
    rows = []; agg = {"downs": Counter(), "boss_dmg": Counter(), "pc_deaths": Counter(),
                      "psaves": Counter(), "vuln": 0, "name_it": 0, "controlled": 0,
                      "calm": Counter(), "calm_lock": 0, "self": 0, "turns": 0, "bunched": 0, "sigh_n": 0, "sigh_hit": 0, "move": 0, "pours": 0, "pulls": 0, "surr": 0, "esc": 0}
    for _ in range(n):
        STATS = new_stats(); H.STATS = STATS; H.TRACK = H.new_track()
        res = fight()
        cat, th, tr, dr = H.score(res)
        res["m"] = H.measures(res)
        res.update(cat=cat, threat=th, triumph=tr, drama=dr, dropped=bool(sum(STATS["downs"].values())))
        rows.append(res)
        agg["downs"].update(STATS["downs"]); agg["boss_dmg"].update(STATS["boss_dmg"])
        for k in res["deaths"]: agg["pc_deaths"][k] += 1
        agg["psaves"].update(STATS["psaves"]); agg["vuln"] += STATS["vuln_hits"]
        agg["name_it"] += STATS["name_it"]; agg["controlled"] += STATS["controlled"]
        agg["calm"].update(STATS["calm"]); agg["calm_lock"] += STATS["calm_lock"]; agg["self"] += STATS["self_hits"]
        for k, k2 in (("turns", "turns"), ("bunched", "bunched"), ("sigh_n", "sigh_n"), ("sigh_hit", "sigh_hit"), ("move", "move_actions"), ("pours", "pours"), ("pulls", "pulls"), ("surr", "surrounds"), ("esc", "escapes")): agg[k] += STATS[k2]
    return rows, agg

def summary(rows, agg):
    n = len(rows); c = Counter(r["cat"] for r in rows)
    wins = [r for r in rows if r["result"] == "win"]
    rounds = [r["p1"] for r in rows if 0 < r["p1"] < 99]
    ps = agg["psaves"]; tot = sum(ps.values()) or 1
    bd = agg["boss_dmg"]; btot = sum(bd.values()) or 1
    return dict(
        win=len(wins) / n, loss=c["Loss"] / n, cruise=c["Cruise"] / n, comfy=c["Comfortable"] / n,
        tense=c["Tense"] / n, pyrrhic=c["Pyrrhic"] / n,
        great=sum(1 for r in wins if not r["deaths"] and r["threat"] >= 4 and r["triumph"] >= 4) / n,
        flat=sum(1 for r in wins if r["threat"] < 2 or r["triumph"] < 2) / n,
        drama=st.mean(r["drama"] for r in rows), rounds=st.mean(rounds) if rounds else float("nan"),
        save_ok=(ps[2] + ps[3]) / tot, saves_per=tot / n,
        sin_share=(bd["Vuln"] + bd["Name It"] + bd["Vuln self"] + bd["Vuln surround"]) / btot,
        ctrl=agg["controlled"] / n, lock=agg["calm_lock"] / n,
        selfhits=agg["self"] / n, bunched=agg["bunched"] / max(1, agg["turns"]),
        sighs=agg["sigh_n"] / n, sigh_hit=agg["sigh_hit"] / max(1, agg["sigh_n"]), move=agg["move"] / n, pours=agg["pours"] / n, pulls=agg["pulls"] / n, surr=agg["surr"] / n, esc=agg["esc"] / n, stale=sum(1 for r in rows if r["result"] == "stalemate") / n)

def compare(n, over):
    hdr = f"{'strategy':12s} {'win':>4s} {'loss':>5s} {'cruise':>6s} {'tense':>5s} {'pyrr':>5s} {'great':>5s} {'flat':>5s} {'drama':>5s} {'rnds':>5s} {'saves':>5s} {'made':>5s} {'sin%':>5s} {'ctrl':>5s}"
    print(hdr)
    out = {}
    for s in STRATEGIES:
        setcfg(dict(over, strategy=s))
        rows, agg = run_many(n); m = summary(rows, agg); out[s] = m
        print(f"{s:12s} {m['win']:4.0%} {m['loss']:5.1%} {m['cruise']:6.0%} {m['tense']:5.0%} {m['pyrrhic']:5.1%} "
              f"{m['great']:5.0%} {m['flat']:5.0%} {m['drama']:5.1f} {m['rounds']:5.1f} {m['saves_per']:5.1f} "
              f"{m['save_ok']:5.0%} {m['sin_share']:5.0%} {m['ctrl']:5.2f}")
    return out

def parse_sets(items):
    out = {}
    for item in items or []:
        k, _, v = item.partition("=")
        if k not in DEFAULTS: raise SystemExit(f"unknown knob {k}")
        try: out[k] = ast.literal_eval(v)
        except (ValueError, SyntaxError): out[k] = v
    return out

def ythrel_spec():
    dmg = CFG["dmg_bonus"]
    vuln = [("Surrounded", f"{CFG['surround_dice']}d6"), ("She starts it", f"{CFG['self_trigger']}d6"),
            ("Flat refusal (crit save)", f"{CFG['vuln_c']}d6")]
    return dict(name="Ythrel, The Soft Collapse", level=16, hp=CFG["ythrel_hp"], ac=CFG["ythrel_ac"],
                perception=28, fort=CFG["ythrel_fort"], ref=26, will=30,
                strikes=[("Claw", 32 + CFG["atk_bonus"], f"3d10+{15 + dmg}"), ("Tail", 30 + CFG["atk_bonus"], f"3d10+{15 + dmg}"),
                         ("Jaws", 32 + CFG["atk_bonus"], f"3d12+{15 + dmg}")],
                dcs=[("Mental (Sigh, There There, Bless)", CFG["dc"])],
                areas=[("Long Sigh", f"{CFG['sigh_dice']}d6", "limited")],
                extra=[(n, dd, "her vulnerability; no table (published sin vulnerabilities 4d6-6d6)") for n, dd in vuln if not dd.startswith("0d")])

def guardrails():
    from pf2e_benchmarks import print_checks, encounter_readout
    print_checks(ythrel_spec())
    encounter_readout(CFG["level"], 3, [16], "Ythrel")

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("-n", type=int, default=4000)
    ap.add_argument("--strategy", default=DEFAULTS["strategy"], choices=STRATEGIES)
    ap.add_argument("--set", action="append")
    ap.add_argument("--seed", type=int)
    ap.add_argument("--log", action="store_true")
    ap.add_argument("--compare", action="store_true")
    ap.add_argument("--winchance", type=int, default=0, metavar="N", help="measure 'in doubt late' with N playouts per round end (slow)")
    a = ap.parse_args()
    if a.seed is not None: R.seed(a.seed)
    global WINCHANCE_N; WINCHANCE_N = a.winchance
    over = parse_sets(a.set)
    if a.compare: compare(a.n, over); return
    setcfg(dict(over, strategy=a.strategy))
    if a.log:
        global STATS
        H.LOGGING = True; STATS = new_stats(); H.STATS = STATS; H.TRACK = H.new_track()
        res = fight(); cat, th, tr, dr = H.score(res)
        print("\n".join(H.LOG)); print(f"\nResult {res['result']} ({cat}) rounds {res['p1']} drama {dr:.1f}")
        return
    guardrails()
    rows, agg = run_many(a.n); m = summary(rows, agg)
    H.show_dashboard(H.dashboard([r["m"] for r in rows]), rounds=m["rounds"],
                     label=f"\nYthrel, strategy '{CFG['strategy']}' ({a.n:,} fights, party level {CFG['level']})")

if __name__ == "__main__":
    main()
