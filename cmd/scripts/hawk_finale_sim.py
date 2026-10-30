"""H.A.W.K. finale simulator: Abyssal Luna (phase 1) -> Avatar of Lilith (phase 2).

Monte Carlo sim of the two-phase boss fight, with both bosses played by the GM
flowchart and the party played by simple scripted turns. Python 3.8+, no
dependencies.

USAGE
  python3 hawk_finale_sim.py                        # recommended setup, 5,000 fights
  python3 hawk_finale_sim.py --preset casual        # another preset
  python3 hawk_finale_sim.py --compare              # every preset side by side
  python3 hawk_finale_sim.py --log --seed 7         # one fight, turn by turn
  python3 hawk_finale_sim.py --set avatar_hp=330 --set last_stand=True
  python3 hawk_finale_sim.py --level 14             # the party's level-14 sheet numbers
  python3 hawk_finale_sim.py --list                 # show presets and every knob

WHAT TO EDIT
  PARTY       per-level character numbers (from the Foundry sheets)
  DEFAULTS    every tuning knob, with a comment on what it does
  PRESETS     named bundles of knob changes

READING THE OUTPUT
  Outcomes    Cruise / Comfortable / Tense / Pyrrhic / Loss
                Cruise       win, no PC ever below half HP
                Comfortable  win, someone below half but none below 25%
                Tense        win, someone below 25% HP or knocked down
                Pyrrhic      win, but a PC died
                Loss         every PC down at once
  Drama       sqrt(Threat x Triumph), 0-10, losses score 0
                Threat   party HP low point, knockdowns, two PCs down at once
                Triumph  60+ damage hits, rescues, the sin turning point, a 35%+ round
  Great       win, nobody dies, Threat >= 4 and Triumph >= 4
  Flat        win where Threat or Triumph stayed under 2

CAVEATS
  Absolute percentages depend on the assumptions in here; comparisons between
  setups are much more reliable than any single number. Not modelled:
  positioning (approximated by probabilities), Arabesque's metamagic and
  defensive spells, Warpwave's random effects, and real player judgement.
"""
import argparse, ast, math, random, statistics as st
from collections import Counter

# ======================================================================= PARTY
# HP, AC, saves and attack numbers per level. Level 14 is from the sheets;
# level 15 is an estimate (master/legendary bumps, greater weapon specialization).
PARTY = {
    14: {
        "Kas":       dict(hp=230, ac=35, fort=23, ref=21, will=22, atk=29, dmg_bonus=9,  class_dc=32),
        "Tavish":    dict(hp=164, ac=36, fort=22, ref=21, will=20, atk=24),
        "Arabesque": dict(hp=104, ac=33, fort=20, ref=22, will=21, dc=33, top_rank=7),
    },
    15: {
        "Kas":       dict(hp=244, ac=36, fort=24, ref=22, will=23, atk=30, dmg_bonus=13, class_dc=33),
        "Tavish":    dict(hp=175, ac=37, fort=23, ref=22, will=21, atk=25),
        "Arabesque": dict(hp=111, ac=34, fort=21, ref=23, will=22, dc=36, top_rank=8),
    },
}

# ==================================================================== DEFAULTS
DEFAULTS = {
    # --- party
    "level": 15,              # which PARTY table to use
    "hero": 1,                # Hero Points each (spent on Heroic Recovery when dying)
    "elixirs": 6,             # Tavish's elixirs of life
    "save_bonus": 0,          # +N to every save for Tavish and Arabesque
    "tav_save": 0,            # +N to Tavish's saves (greater resilient rune = 2)
    "ara_save": 0,            # +N to Arabesque's saves (greater bands of force = 1)
    "ara_ac": 0,              # +N to Arabesque's AC (greater bands of force = 1)
    "haste": True,            # Arabesque opens with haste on Kas and Tavish
    # --- player habits (all False/0 = the scripted "good" play)
    "kas_simple": False,      # Kas only makes plain Strikes
    "no_relics": False,       # Tavish never uses Flare Bolt / Tidal Crash
    "late_heal": False,       # Tavish only heals allies who are already down
    "ara_any_save": False,    # Arabesque's spells target a random save, not Fortitude
    "waste_p": 0.0,           # chance each PC loses an action on a turn
    "heal_thresh": 0.3,       # HP fraction at which PCs drink potions / feed elixirs
    "kas_step": 0.25,         # phase 2: chance Kas body-blocks (an act of loyalty) before the sin is known
    # --- bosses
    "luna_hp": 270,
    "luna_ac": 36,
    "avatar_hp": 300,
    "spores_dice": 15,        # Destiny Spores damage dice (d6)
    "spores_p": 0.65,         # chance Luna uses Spores when it's ready and 2+ PCs are up
    "dive_p": 0.75,           # chance Luna dives Arabesque/Tavish; 'adaptive' = director rule
    "challenge_fate": True,   # Luna's reaction on Kas's first swing each round
    "wail_dice": 8,           # Wail damage dice (d10)
    "wail_p": 0.75,           # chance the avatar Wails when it's ready
    "wail_mercy": False,      # director rule: the Wail never targets a PC below 40% HP
    "wrath_p": 0.5,           # chance the avatar casts Divine Wrath when 2+ PCs are up
    "dispel": True,           # avatar dispels haste; 'adaptive' = only when the party struggles
    "song_at": 0.5,           # avatar sings below this HP fraction (1.0 = opens with it, 0 = never)
    "ara_weight": 2,          # how often the avatar picks Arabesque for melee (Kas = 6, Tavish = 2)
    "last_stand": False,      # avatar's once-only Last Stand at 25% HP
    "last_stand_heal": 40,    # HP she regains on Last Stand
    # --- the sin of betrayal
    "sin_offset": 12,         # flat add on the 4d6 sin damage (12 cancels her resistance)
    "disc": None,             # {round: chance} the party works out the sin; None = default curve
    "disc_late": 1.0,         # chance per round after the curve ends
    # --- transformation
    "transition_heal": 0,     # HP everyone regains when Luna falls
    "heal_cond": False,       # only give that heal if the party is struggling
    # --- what-ifs
    "no_party_crits": False,  # remove every crit that favours the party
    "no_boss_crits": False,   # remove every crit that favours the bosses
}

RECOMMENDED = dict(tav_save=2, ara_save=1, ara_ac=1, dispel="adaptive", dive_p="adaptive",
                   wail_mercy=True, transition_heal=75, heal_cond=True)
PRESETS = {
    "as_built":       dict(),
    "recommended":    dict(RECOMMENDED),
    "last_stand":     dict(RECOMMENDED, last_stand=True),
    "casual":         dict(RECOMMENDED, ara_any_save=True, kas_simple=True, no_relics=True,
                           late_heal=True, waste_p=0.33),
    "casual_no_haste": dict(RECOMMENDED, ara_any_save=True, kas_simple=True, no_relics=True,
                           late_heal=True, waste_p=0.33, haste=False),
    "defensive":      dict(RECOMMENDED, spores_p=0.35, wail_p=0.45, wrath_p=0.25, dive_p=0.35,
                           ara_ac=3, ara_weight=1, heal_thresh=0.5, kas_step=0.5, waste_p=0.25),
}

# ===================================================================== engine
R = random.Random()
CFG = dict(DEFAULTS)
STATS = None      # per-fight counters
TRACK = None      # per-fight drama/tension tracking
LOG = []
LOGGING = False
STRUG = False

def d(n, s): return sum(R.randint(1, s) for _ in range(n))

def log(s):
    if LOGGING: LOG.append(s)

def degree(roll, total, dc):
    """0 crit fail, 1 fail, 2 success, 3 crit success."""
    g = 3 if total >= dc + 10 else 2 if total >= dc else 0 if total <= dc - 10 else 1
    if roll == 20: g = min(3, g + 1)
    if roll == 1: g = max(0, g - 1)
    return g

def adj(kind, g):
    """Apply the no-crit what-ifs and count every roll's outcome."""
    np, nb = CFG["no_party_crits"], CFG["no_boss_crits"]
    if kind == "pc_att":
        if np and g == 3: g = 2
        if nb and g == 0: g = 1
    elif kind == "boss_att":
        if nb and g == 3: g = 2
        if np and g == 0: g = 1
    elif kind == "pc_save":
        if nb and g == 0: g = 1
        if np and g == 3: g = 2
    elif kind == "boss_save":
        if np and g == 0: g = 1
        if nb and g == 3: g = 2
    STATS["rolls"][(kind, g)] += 1
    return g

def new_stats():
    return {"downs": Counter(), "deaths": Counter(), "boss_dmg": Counter(), "rolls": Counter(),
            "song": Counter(), "heroic": 0, "stun4": 0, "sin_triggers": 0}

def new_track():
    return {"low": 1.0, "pc_low": {}, "brink": False, "rescues": 0, "big": 0,
            "holy_hit": False, "comeback": False, "party": None}

# ----------------------------------------------------------------------- PCs
class PC:
    def __init__(s, name, hp, ac, fort, ref, will, **kw):
        s.name, s.maxhp, s.hp, s.ac = name, hp, hp, ac
        s.sv = {"fort": fort, "ref": ref, "will": will}
        s.temp = 0; s.frightened = 0; s.slowed = 0; s.stunned = 0
        s.doomed = 0; s.wounded = 0; s.dying = 0; s.dead = False
        s.sickened = 0; s.lost_turn = 0; s.blinded = 0; s.stupefied = (0, 0)
        s.hasted = False; s.holy = False; s.potions = []; s.illomen = 0
        s.hero = CFG["hero"]; s.stable = False
        s.__dict__.update(kw)

    @property
    def up(s): return s.hp > 0 and not s.dead

    def save(s, kind, dc):
        mod = s.sv[kind] - s.frightened - s.sickened
        r = R.randint(1, 20)
        return adj("pc_save", degree(r, r + mod, dc))

    def take(s, dmg, src, crit=False):
        if s.dead or dmg <= 0: return
        a = min(s.temp, dmg); s.temp -= a; dmg -= a
        if s.hp <= 0:  # already down
            if s.stable: s.stable = False; s.dying = 1 + s.wounded
            s.dying += 2 if crit else 1; s.check_death(src)
        else:
            s.hp -= dmg
            if s.hp <= 0:
                s.hp = 0; s.dying = 1 + (1 if crit else 0) + s.wounded
                STATS["downs"][src] += 1; log(f"   >> {s.name} DOWN ({src}), dying {s.dying}")
                s.check_death(src)
        track_party(s)

    def check_death(s, src):
        limit = (5 if s.name == "Kas" else 4) - s.doomed   # Kas has Inextinguishable
        if s.dying >= limit and not s.dead and s.hero > 0:
            s.hero = 0; s.dying = 0; s.stable = True; STATS["heroic"] += 1
            log(f"   !! {s.name} spends Hero Points: Heroic Recovery, unconscious but stable"); return
        if s.dying >= limit and not s.dead:
            s.dead = True; STATS["deaths"][src] += 1; log(f"   XX {s.name} DIES ({src})")

    def heal(s, amt):
        if s.dead: return
        was_down = s.hp <= 0
        if was_down and amt > 0:
            s.wounded += 1; s.dying = 0; s.stable = False
        s.hp = min(s.maxhp, s.hp + amt)
        if was_down and s.hp > 0: TRACK["rescues"] += 1

def track_party(pc):
    party = TRACK["party"]
    TRACK["pc_low"][pc.name] = min(TRACK["pc_low"].get(pc.name, 1.0), max(pc.hp, 0) / pc.maxhp)
    frac = sum(max(p.hp, 0) for p in party) / sum(p.maxhp for p in party)
    TRACK["low"] = min(TRACK["low"], frac)
    if sum(1 for p in party if p.hp <= 0) >= 2: TRACK["brink"] = True

def make_party():
    P = PARTY[CFG["level"]]
    tb = CFG["save_bonus"] + CFG["tav_save"]; ab = CFG["save_bonus"] + CFG["ara_save"]
    k, t, a = P["Kas"], P["Tavish"], P["Arabesque"]
    kas = PC("Kas", k["hp"], k["ac"], k["fort"], k["ref"], k["will"], atk=k["atk"], bravery=True,
             dmg_bonus=k["dmg_bonus"], class_dc=k["class_dc"], potions=[47, 24], wave=True,
             lvl=CFG["level"])
    tav = PC("Tavish", t["hp"], t["ac"], t["fort"] + tb, t["ref"] + tb, t["will"] + tb,
             atk=t["atk"], elixirs=CFG["elixirs"], potions=[24, 24], tidal=True)
    top = a["top_rank"]
    ara = PC("Arabesque", a["hp"], a["ac"] + CFG["ara_ac"], a["fort"] + ab, a["ref"] + ab, a["will"] + ab,
             dc=a["dc"], slots=[r for r in range(top, 4, -1) for _ in range(4)])
    party = [kas, tav, ara]
    TRACK["party"] = party
    return party

# ---------------------------------------------------------------------- bosses
class Boss:
    def __init__(s, name, hp, ac, fort, ref, will, resist=0, weak=0, fh=0):
        s.name, s.maxhp, s.hp, s.ac = name, hp, hp, ac
        s.sv = {"fort": fort, "ref": ref, "will": will}
        s.resist, s.weak, s.fh = resist, weak, fh
        s.frightened = 0; s.persist = 0; s.reaction = True
        s.clumsy = 0; s.enfeebled = 0; s.prone = False

    def status(s, kind="ac"):
        return max(s.frightened, 2 if (s.clumsy and kind in ("ac", "ref")) else 0,
                   2 if (s.enfeebled and kind == "atk") else 0)
    def eff_ac(s): return s.ac - s.status("ac") - (2 if s.prone else 0)
    def atk_pen(s): return s.status("atk") + (2 if s.prone else 0)

    def save(s, kind, dc):
        r = R.randint(1, 20)
        return adj("boss_save", degree(r, r + s.sv[kind] - s.status(kind), dc))

    def take(s, raw, holy=False, src=""):
        if raw <= 0: return 0
        dmg = raw + s.weak if holy else max(0, raw - s.resist)
        s.hp -= dmg; STATS["boss_dmg"][src] += dmg
        if dmg >= 60 and not src.startswith("Sin"): TRACK["big"] += 1
        if "holy" in src and dmg > 0: TRACK["holy_hit"] = True
        return dmg

# ------------------------------------------------------------------- PC turns
def actions_for(pc):
    a = 3 + (1 if pc.hasted else 0)
    if CFG["waste_p"] and R.random() < CFG["waste_p"]: a -= 1
    lose = max(pc.slowed, 0) + min(pc.stunned, 3)
    pc.stunned = max(0, pc.stunned - 3); pc.slowed = 0
    return max(0, a - lose)

def strike(pc, boss, atk_mod, worse=False):
    r = R.randint(1, 20)
    if worse: r = min(r, R.randint(1, 20))
    if pc.blinded and R.random() < 0.5: return 0, None
    tot = r + atk_mod - pc.frightened - pc.sickened
    return adj("pc_att", degree(r, tot, boss.eff_ac())), r

def kas_hit(kas, boss, ctx, mod, extra_dice, label, worse=False, brutal=False):
    g, _ = strike(kas, boss, mod, worse)
    holy = kas.holy and g >= 2
    if g >= 2:
        dmg = d(3 + extra_dice, 12) + kas.dmg_bonus
        if g == 3: dmg *= 2
        dealt = boss.take(dmg, holy=holy, src="Kas hammer" + (" (holy)" if holy else ""))
        if holy: kas.holy = False
        note = ""
        if g == 3:   # crit package: clumsy, enfeebled, frightened 2, prone on a failed Fort
            boss.frightened = max(boss.frightened, 2); boss.clumsy = 1; boss.enfeebled = 1
            if boss.save("fort", kas.class_dc) <= 1: boss.prone = True; note = ", knocked PRONE"
            note = ", clumsy/enfeebled/frightened 2" + note
        log(f"   Kas {label} {'CRITS' if g == 3 else 'hits'} for {dealt}{' holy' if holy else ''}{note}")
        if g == 3 and ctx["phase"] == 2 and boss.reaction: vengeful_spite(boss, kas, ctx)
    elif brutal and g == 1:
        dealt = boss.take(d(1, 12), src="Kas hammer"); log(f"   Kas {label} misses but still deals {dealt}")
    else:
        log(f"   Kas {label} misses" + (" (Challenge Fate)" if worse else ""))
    return g

def kas_turn(kas, boss, ctx):
    acts = actions_for(kas); party = ctx["party"]
    if kas.lost_turn: kas.lost_turn -= 1; log("   Kas is confused and loses his turn"); return
    hurt = [p for p in party if not p.dead and (p.hp <= 0 or p.hp < 0.5 * p.maxhp)]
    if kas.wave and acts >= 2 and (len(hurt) >= 2 or any(p.hp <= 0 and not p.dead for p in party)):
        kas.wave = False; acts -= 2
        for p in party:
            if p.dead: continue
            if p is kas or R.random() < 0.7: p.heal(d(kas.lvl, 4))
        log(f"   Kas activates HEALING WAVE ({kas.lvl}d4 to himself and allies in the cone)")
        if ctx["phase"] == 2: sin_trigger(kas, kas, boss, ctx, "heals his allies with Healing Wave")
    elif kas.hp < CFG["heal_thresh"] * kas.maxhp and kas.potions and acts:
        h = kas.potions.pop(0); kas.heal(h); acts -= 1; log(f"   Kas drinks a potion (+{h})")
    if ctx["phase"] == 2 and ctx["discovered"] and acts >= 3:
        sin_trigger(kas, kas, boss, ctx, "Swear Fidelity"); acts -= 1
    elif ctx["phase"] == 2 and not ctx["discovered"] and R.random() < CFG["kas_step"] and acts >= 3:
        sin_trigger(kas, kas, boss, ctx, "steps between the avatar and an ally"); acts -= 1
    worse = ctx["phase"] == 1 and boss.reaction and CFG["challenge_fate"]
    if kas.illomen: worse = True; kas.illomen -= 1
    map_ = 0
    need_move = R.random() < ctx["kas_move"]
    if need_move and acts >= 2:                       # Sudden Charge: two Strides + Strike
        if worse and ctx["phase"] == 1: boss.reaction = False
        kas_hit(kas, boss, ctx, kas.atk, 0, "Sudden Charge", worse); acts -= 2; map_ = 1; worse = False
    elif need_move and acts:
        acts -= 1; log("   Kas closes the distance")
    if acts >= 2 and map_ == 0 and boss.hp > 0:       # Vicious Swing: +2 dice, counts as two attacks
        if worse and ctx["phase"] == 1: boss.reaction = False
        if CFG["kas_simple"]:
            kas_hit(kas, boss, ctx, kas.atk, 0, "Strike", worse); acts -= 1; map_ = 1; worse = False
        else:
            kas_hit(kas, boss, ctx, kas.atk, 2, "Vicious Swing", worse); acts -= 2; map_ = 2; worse = False
    while acts > 1 and boss.hp > 0 and kas.hasted:    # hasted Strike before the finisher
        kas_hit(kas, boss, ctx, kas.atk - min(10, 5 * map_), 0, "Strike"); acts -= 1; map_ += 1
    if acts >= 1 and boss.hp > 0:                     # Brutal Finish ends the turn
        if worse and ctx["phase"] == 1: boss.reaction = False
        if CFG["kas_simple"]:
            while acts > 0 and boss.hp > 0:
                kas_hit(kas, boss, ctx, kas.atk - min(10, 5 * map_), 0, "Strike", worse)
                acts -= 1; map_ += 1; worse = False
        else:
            kas_hit(kas, boss, ctx, kas.atk - min(10, 5 * map_), 1, "Brutal Finish", worse, brutal=True)

def tav_turn(tav, party, boss, ctx):
    acts = actions_for(tav)
    if tav.lost_turn: tav.lost_turn -= 1; log("   Tavish is confused and loses his turn"); return
    tgt = next((p for p in party if p is not tav and not p.dead and
                (p.hp <= 0 or (not CFG["late_heal"] and p.hp < CFG["heal_thresh"] * p.maxhp))), None)
    if tgt and tav.elixirs and acts >= 2:
        h = d(5, 6) + 25; tav.elixirs -= 1; tgt.heal(h); acts -= 2
        log(f"   Tavish feeds {tgt.name} an elixir of life (+{h})")
        if ctx["phase"] == 2: sin_trigger(tav, tgt, boss, ctx, "hands an ally an elixir")
    elif tav.hp < CFG["heal_thresh"] * tav.maxhp and tav.potions and acts:
        h = tav.potions.pop(0); tav.heal(h); acts -= 1; log(f"   Tavish drinks a potion (+{h})")
    if ctx["phase"] == 2 and ctx["discovered"] and acts >= 2 and not tav.holy:
        sin_trigger(tav, tav, boss, ctx, "Swear Fidelity"); acts -= 1
    n = 0
    while acts > 0 and boss.hp > 0:
        mod = tav.atk - (0 if n == 0 else 5 if n == 1 else 10)
        relic = None
        if n == 0 and acts >= 2 and not CFG["no_relics"]:
            relic = "tidal" if (ctx["phase"] == 2 and tav.tidal) else "flare"
        g, _ = strike(tav, boss, mod)
        holy = tav.holy
        dmg = {3: d(3, 6) * 2 + 5, 2: d(3, 6) + 5, 1: 5, 0: 0}[g]   # greater bomb + splash
        if relic and g >= 2:
            dmg += (d(8, 10) if relic == "tidal" else d(7, 6)) * (2 if g == 3 else 1)
        if relic == "tidal": tav.tidal = False
        dealt = boss.take(dmg, holy=holy and g >= 1, src="Tavish bombs" + (" (holy)" if holy else ""))
        if holy and g >= 1: tav.holy = False
        if g >= 2 and ctx["phase"] == 1: boss.persist = 5          # Sticky Bomb
        name = {"tidal": "TIDAL CRASH bomb", "flare": "Flare Bolt bomb", None: "bomb"}[relic]
        log(f"   Tavish {name} {['crit-misses', 'misses (splash)', 'hits', 'CRITS'][g]} for {dealt}{' holy' if holy else ''}")
        acts -= 2 if relic else 1; n += 1

def ara_turn(ara, party, boss, ctx):
    acts = actions_for(ara)
    if ara.lost_turn: ara.lost_turn -= 1; log("   Arabesque is confused and loses her turn"); return
    if CFG["haste"] and ctx["phase"] == 1 and not ctx["haste_cast"] and acts >= 2:
        ctx["haste_cast"] = True; acts -= 2
        for p in party: p.hasted = p.up and p.name != "Arabesque"
        log("   Arabesque casts 7th-rank haste on Kas and Tavish")
    elif ctx["phase"] == 2 and ctx["haste_dispelled"] and acts >= 2 and R.random() < 0.5:
        ctx["haste_dispelled"] = False; acts -= 2
        for p in party: p.hasted = p.up and p.name != "Arabesque"
        log("   Arabesque re-casts haste on her allies")
        sin_trigger(ara, party[0], boss, ctx, "casts haste on allies")
    if ctx["phase"] == 2 and ctx["discovered"] and acts >= 3 and not ara.holy:
        sin_trigger(ara, ara, boss, ctx, "Swear Fidelity"); acts -= 1
    if acts >= 2 and ara.slots and boss.hp > 0:
        rank = ara.slots.pop(0); acts -= 2
        st_val, st_rounds = ara.stupefied
        if st_rounds and R.randint(1, 20) < 5 + st_val:
            log("   Arabesque's spell is lost to stupefied"); return
        dmg = d(2 * rank, 6)                                     # vampiric exsanguination
        kind = R.choice(["fort", "ref", "will"]) if CFG["ara_any_save"] else "fort"
        g = boss.save(kind, ara.dc - ara.frightened - ara.sickened - st_val)
        mult = [2, 1, 0.5, 0][g]
        holy = ara.holy
        dealt = boss.take(int(dmg * mult), holy=holy and mult > 0, src="Arabesque spells" + (" (holy)" if holy else ""))
        if holy and mult > 0: ara.holy = False
        ara.temp = max(ara.temp, dealt // 2)
        log(f"   Arabesque casts {rank}th-rank vampiric exsanguination: "
            f"{['crit fail', 'fail', 'success', 'crit success'][g]}, {dealt} dmg{' holy' if holy else ''} (+{dealt // 2} temp HP)")
    if ara.hp < max(0.4, CFG["heal_thresh"]) * ara.maxhp and acts and party[1].potions:
        h = party[1].potions.pop(0); ara.heal(h); log(f"   Arabesque drinks a potion Tavish gave her earlier (+{h})")

# -------------------------------------------------------------------- the sin
def sin_trigger(actor, beneficiary, boss, ctx, how):
    if actor.name in ctx["sin_used"]: return            # once per creature per round
    ctx["sin_used"].add(actor.name); ctx["triggered_this_round"] = True
    dmg = boss.take(d(4, 6) + CFG["sin_offset"], src="Sin of Betrayal")
    beneficiary.holy = True
    STATS["sin_triggers"] += 1
    log(f"   * {actor.name} {how}: SIN triggers, avatar takes {dmg}; {beneficiary.name}'s next attack is holy")

# ----------------------------------------------------------------- boss turns
def struggling(party):
    live = [p for p in party if not p.dead]
    if any(p.hp <= 0 for p in live): return True
    return sum(max(p.hp, 0) for p in live) < 0.5 * sum(p.maxhp for p in live)

def pick(party, weights):
    up = [p for p in party if p.up]
    w = [weights.get(p.name, 1) for p in up]
    if not up or sum(w) == 0: return None
    return R.choices(up, w)[0]

def boss_strike(boss, tgt, mod, dmgfn, label, ctx, extra=None):
    if tgt is None or not tgt.up: return 0
    r = R.randint(1, 20); tot = r + mod - boss.atk_pen()
    g = adj("boss_att", degree(r, tot, tgt.ac - tgt.frightened - tgt.sickened))
    if g >= 2:
        dmg = max(1, dmgfn() - (2 if boss.enfeebled else 0)) * (2 if g == 3 else 1)
        tgt.take(dmg, label, crit=(g == 3) and tgt.hp <= dmg)
        log(f"   {label} {'CRITS' if g == 3 else 'hits'} {tgt.name} for {dmg} ({max(tgt.hp, 0)}/{tgt.maxhp})")
        if extra: extra(tgt, g)
    else:
        log(f"   {label} misses {tgt.name}")
    return g

def basic_save_dmg(tgt, kind, dc, dmg, label, on_fail=None):
    g = tgt.save(kind, dc)
    mult = [2, 1, 0.5, 0][g]
    tgt.take(int(dmg * mult), label, crit=(g == 0))
    log(f"     {tgt.name}: {['CRIT FAIL', 'fail', 'success', 'crit success'][g]} -> {int(dmg * mult)} ({max(tgt.hp, 0)}/{tgt.maxhp})")
    if on_fail and g <= 1 and tgt.up: on_fail(tgt, g)
    return g

def luna_turn(luna, party, ctx):
    luna.reaction = True
    was_prone = luna.prone; luna.prone = False; luna.clumsy = 0; luna.enfeebled = 0
    luna.frightened = max(0, luna.frightened - 1)
    if luna.persist:
        luna.take(luna.persist, src="Tavish bombs"); log(f"   Luna takes {luna.persist} persistent")
        if R.randint(1, 20) >= 15: luna.persist = 0
    up = [p for p in party if p.up]
    if not up: return
    def impending(t, g):
        if g == 3:
            t.doomed += 1; ctx["spores_cd"] = 0
            log(f"     Impending Fate: {t.name} doomed {t.doomed}; Destiny Spores recharges")
    # 1. Destiny Spores
    if ctx["spores_cd"] == 0 and len(up) >= 2 and R.random() < CFG["spores_p"]:
        tg = R.sample(up, 3 if len(up) == 3 and R.random() < 0.4 else 2)
        dmg = d(CFG["spores_dice"], 6)
        log(f"  Luna: DESTINY SPORES ({dmg} mental) on {', '.join(t.name for t in tg)}, then flies off")
        def spore_fail(t, g):
            t.slowed = 2 if g == 0 else 1
            if g == 0: t.doomed += 1; log(f"     {t.name} slowed 2 and doomed {t.doomed}")
        for t in tg: basic_save_dmg(t, "will", 39, dmg, "Destiny Spores", spore_fail)
        ctx["spores_cd"] = R.randint(1, 4); ctx["kas_move"] = 0.8
        return
    # 2. Dive the squishies with Abyssal Frenzy
    dive = CFG["dive_p"]
    if dive == "adaptive": dive = 0.25 if struggling(party) else 1.0
    if R.random() < dive and not was_prone:
        t = pick(party, {"Arabesque": 55, "Tavish": 45, "Kas": 0}) or pick(party, {})
        log(f"  Luna: flies in, ABYSSAL FRENZY on {t.name}")
        boss_strike(luna, t, 28, lambda: d(3, 8) + 14 + d(1, 8), "Stab", ctx, impending)
        for m in (26, 22):
            if t.up: boss_strike(luna, t, m, lambda: d(3, 6) + 14 + d(1, 8), "Slash", ctx, impending)
        ctx["kas_move"] = 0.6
        return
    kas = party[0]
    # 3. Walk the Timelines (once)
    if not ctx["walked"]:
        ctx["walked"] = True
        log("  Luna: WALK THE TIMELINES (strikes Kas and a squishy), flies clear")
        boss_strike(luna, kas if kas.up else pick(party, {}), 28, lambda: d(3, 8) + 14 + d(1, 8), "Stab (copy 1)", ctx, impending)
        t = pick(party, {"Arabesque": 1, "Tavish": 1, "Kas": 0}) or pick(party, {})
        boss_strike(luna, t, 28, lambda: d(3, 8) + 14 + d(1, 8), "Stab (copy 2)", ctx, impending)
        ctx["kas_move"] = 0.8
        return
    # 4. Ill Omen + Slash
    t = kas if kas.up else pick(party, {})
    g = t.save("will", 39)
    if g <= 1: t.illomen = 3 if g == 0 else 1
    log(f"  Luna: ILL OMEN on {t.name} ({['crit fail', 'fail', 'success', 'crit success'][g]}) + Slash")
    boss_strike(luna, t, 30, lambda: d(3, 6) + 14 + d(1, 8), "Slash", ctx, impending)
    ctx["kas_move"] = 0.2

def terrifying(t, g_hit, dc=38):
    if t.frightened or not t.up: return
    g = t.save("will", dc + (5 if g_hit == 3 else 0))
    brave = getattr(t, "bravery", False)
    if brave and g == 2: g = 3
    if g == 1: t.frightened = 2 - (1 if brave else 0)
    if g == 0:
        t.frightened = 2 - (1 if brave else 0); t.stunned += 4; STATS["stun4"] += 1
    if g <= 1: log(f"     Terrifying Touch: {t.name} frightened {t.frightened}" + (" and STUNNED 4" if g == 0 else ""))

def vengeful_spite(av, t, ctx):
    av.reaction = False
    dmg = d(4, 10) + 14; log(f"   Avatar reacts: VENGEFUL SPITE on {t.name} ({dmg} mental)")
    g = basic_save_dmg(t, "will", 38, dmg, "Vengeful Spite")
    if t.up and not t.frightened and g <= 1:
        t.frightened = 2 - (1 if getattr(t, "bravery", False) else 0)
        if g == 0: t.stunned += 4; STATS["stun4"] += 1

def song(party, ctx):
    for t in party:
        if not t.up or t.name in ctx["song_immune"]: continue
        g = t.save("will", 38)
        if g == 3: ctx["song_immune"].add(t.name); continue
        if g == 2: continue
        roll = R.randint(1, 4) + (1 if g == 0 else 0)
        eff = {1: "frightened 2", 2: "confused", 3: "stupefied 4", 4: "blinded", 5: "STUNNED + stupefied 1"}[roll]
        if roll == 1: t.frightened = max(t.frightened, 2 - (1 if getattr(t, "bravery", False) else 0))
        if roll in (2, 5): t.lost_turn = 1
        if roll == 3: t.stupefied = (4, 1)
        if roll == 4: t.blinded = 1
        STATS["song"][eff] += 1
        log(f"     Song: {t.name} {['crit fails', 'fails'][g]}, {eff}")

def avatar_turn(av, party, ctx):
    av.reaction = True
    prone_cost = 1 if av.prone else 0
    av.prone = False; av.clumsy = 0; av.enfeebled = 0
    av.frightened = max(0, av.frightened - 1)
    if av.hp < av.maxhp: av.hp = min(av.maxhp, av.hp + av.fh)          # fast healing
    up = [p for p in party if p.up]
    if not up: return
    acts = 3 - prone_cost
    if CFG["last_stand"] and not ctx.get("last") and av.hp <= av.maxhp * 0.25:
        ctx["last"] = True; ctx["wail_cd"] = 0; acts += 1; av.hp += CFG["last_stand_heal"]
        log(f"  Avatar: LAST STAND (+{CFG['last_stand_heal']} HP, the Wail recharges, +1 action)")
    song_at = CFG["song_at"]
    if (av.hp < av.maxhp * song_at or song_at >= 1) and not ctx["song"]:
        ctx["song"] = True; acts -= 2; log("  Avatar: UNFATHOMABLE SONG on everyone"); song(party, ctx)
    elif ctx["song"]:
        acts -= 1; log("  Avatar: sustains Unfathomable Song"); song(party, ctx)
    weights = {"Kas": 6, "Tavish": 2, "Arabesque": CFG["ara_weight"]}
    slash = lambda: d(3, 6) + 17
    # 1. Wail
    wail_pool = [p for p in up if not (CFG["wail_mercy"] and p.hp < 0.4 * p.maxhp)]
    if ctx["wail_cd"] == 0 and len(wail_pool) >= 2 and R.random() < CFG["wail_p"]:
        tg = R.sample(wail_pool, 3 if len(wail_pool) == 3 and R.random() < 0.5 else 2)
        dmg = d(CFG["wail_dice"], 10)
        log(f"  Avatar: WAIL ({dmg} spirit) on {', '.join(t.name for t in tg)}")
        for t in tg: basic_save_dmg(t, "fort", 38, dmg, "Wail")
        ctx["wail_cd"] = R.randint(1, 4); acts -= 2
        if acts and any(p.up for p in party): boss_strike(av, pick(party, weights), 32, slash, "Slash", ctx)
        return
    # 2. Dispel haste
    want = CFG["dispel"]
    if want == "adaptive": want = struggling(party)
    if want and any(p.hasted for p in party if p.up) and acts >= 2:
        acts -= 2
        if R.randint(1, 20) + 28 - av.frightened >= 36:
            for p in party: p.hasted = False
            ctx["haste_dispelled"] = True; log("  Avatar: DISPEL MAGIC strips haste")
        else:
            log("  Avatar: Dispel Magic fails against haste")
        if acts: boss_strike(av, pick(party, {"Kas": 4, "Tavish": 3, "Arabesque": 3}), 30, lambda: d(2, 10) + 17, "Stinger", ctx)
        return
    # 3. Divine Wrath
    if len(up) >= 2 and R.random() < CFG["wrath_p"] and acts >= 2:
        tg = up if len(up) == 2 or R.random() < 0.5 else R.sample(up, 2)
        dmg = d(9, 10); log(f"  Avatar: DIVINE WRATH ({dmg} spirit) on {', '.join(t.name for t in tg)}")
        for t in tg:
            g = t.save("fort", 38)
            m = [1, 1, 0.5, 0][g]; t.take(int(dmg * m), "Divine Wrath")
            if g <= 1 and t.up: t.sickened = 2 if g == 0 else 1
            if g == 0 and t.up: t.slowed = max(t.slowed, 1)
            log(f"     {t.name}: {['CRIT FAIL', 'fail', 'success', 'crit success'][g]} -> {int(dmg * m)} ({max(t.hp, 0)}/{t.maxhp})")
        acts -= 2
        if acts and any(p.up for p in party): boss_strike(av, pick(party, weights), 32, slash, "Slash", ctx)
        return
    # 4. Demonic Frenzy + Hand of Chaos
    t = pick(party, weights)
    if acts >= 2:
        log(f"  Avatar: DEMONIC FRENZY on {t.name}")
        boss_strike(av, t, 30, lambda: d(2, 10) + 17, "Stinger", ctx)
        if t.up: boss_strike(av, t, 28, slash, "Slash", ctx)
        if t.up: boss_strike(av, t, 24, slash, "Slash", ctx)
        acts -= 2
    if acts and t and t.up:
        boss_strike(av, t, 22, lambda: d(4, 10) + 14, "Hand of Chaos", ctx, terrifying)

# ------------------------------------------------------------------- the fight
def end_of_turn(p):
    p.frightened = max(0, p.frightened - 1)
    if p.blinded: p.blinded -= 1
    v, r = p.stupefied
    if r: p.stupefied = (v, r - 1) if r > 1 else (0, 0)

def dying_check(p):
    if p.dead or p.hp > 0 or p.stable: return
    r = R.randint(1, 20); g = degree(r, r, 10 + p.dying)
    p.dying += {0: 2, 1: 1, 2: -1, 3: -2}[g]
    if p.dying <= 0:
        p.dying = 0; p.hp = 1; p.wounded += 1; log(f"   {p.name} stabilises and wakes at 1 HP")
    else:
        p.check_death("bled out (recovery checks)")

def run_phase(boss, party, phase, ctx, max_rounds=15):
    order = party + [boss]; R.shuffle(order)
    prev_hp = None
    for rnd in range(1, max_rounds + 1):
        if prev_hp is not None and prev_hp - max(boss.hp, 0) >= 0.35 * boss.maxhp: TRACK["comeback"] = True
        prev_hp = max(boss.hp, 0)
        ctx["round"] = rnd; ctx["sin_used"] = set(); ctx["triggered_this_round"] = False
        log(f"-- Phase {phase}, round {rnd}: {boss.name} {max(boss.hp, 0)}/{boss.maxhp} | " +
            " | ".join(f"{p.name} {'DEAD' if p.dead else max(p.hp, 0)}" + (f" d{p.doomed}" if p.doomed else "") for p in party))
        for c in order:
            if boss.hp <= 0: return rnd
            if not any(p.up for p in party): return -rnd
            if c is boss:
                (luna_turn if phase == 1 else avatar_turn)(boss, party, ctx)
                if phase == 1 and ctx["spores_cd"]: ctx["spores_cd"] -= 1
                if phase == 2 and ctx["wail_cd"]: ctx["wail_cd"] -= 1
            else:
                if c.dead: continue
                if c.hp <= 0: dying_check(c); continue
                log(f"  {c.name}:")
                if c.name == "Kas": kas_turn(c, boss, ctx)
                elif c.name == "Tavish": tav_turn(c, party, boss, ctx)
                else: ara_turn(c, party, boss, ctx)
                end_of_turn(c)
        if boss.hp <= 0: return rnd
        if not any(p.up for p in party): return -rnd
        if phase == 2 and not ctx["discovered"]:     # does the party work out the sin?
            curve = CFG["disc"]
            if curve is None:
                p = {1: 0.0, 2: 0.25, 3: 0.5, 4: 0.6}.get(rnd, CFG["disc_late"])
                if rnd < 3 and not ctx["triggered_this_round"]: p = 0
            else:
                p = curve.get(rnd, CFG["disc_late"])
            if R.random() < p:
                ctx["discovered"] = True; ctx["discovered_round"] = rnd
                log(f"  ** The party works out the sin (end of round {rnd})")
    return 99

def transition(party):
    global STRUG
    STRUG = struggling(party)
    log("== TRANSFORMATION: free round" + (" (party is struggling)" if STRUG else ""))
    tav = party[1]
    for p in party:
        if p.dead: continue
        if p.hp <= 0:
            h = d(5, 6) + 25 if tav.elixirs else (tav.potions.pop(0) if tav.potions else 0)
            if tav.elixirs: tav.elixirs -= 1
            p.heal(h); log(f"   {p.name} revived (+{h})")
        if p.hp < 0.6 * p.maxhp and p.potions:
            h = p.potions.pop(0); p.heal(h); log(f"   {p.name} drinks (+{h})")
        if CFG["transition_heal"] and p.up and (not CFG["heal_cond"] or STRUG):
            p.heal(CFG["transition_heal"]); log(f"   {p.name} regains {CFG['transition_heal']} HP from the transformation")
        p.frightened = p.slowed = p.stunned = p.sickened = p.lost_turn = p.blinded = 0
        p.stupefied = (0, 0)

def fight():
    """One full fight. Uses the global STATS and TRACK, which the caller resets."""
    party = make_party()
    ctx = {"party": party, "phase": 1, "spores_cd": 0, "kas_move": 0.5, "walked": False,
           "haste_cast": False, "haste_dispelled": False, "discovered": False,
           "discovered_round": None, "song": False, "song_immune": set(), "wail_cd": 0}
    luna = Boss("Luna", CFG["luna_hp"], CFG["luna_ac"], 27, 28, 29)
    r1 = run_phase(luna, party, 1, ctx)
    res = {"p1": r1, "p2": None, "deaths": [p.name for p in party if p.dead], "doomed_in": {}}
    if r1 < 0:
        res["result"] = "loss"; return res
    transition(party)
    res["doomed_in"] = {p.name: p.doomed for p in party}
    ctx["phase"] = 2; ctx["kas_move"] = 0.2
    av = Boss("Avatar", CFG["avatar_hp"], 39, 25, 29, 32, resist=12, weak=15, fh=10)
    r2 = run_phase(av, party, 2, ctx)
    res["p2"] = r2; res["deaths"] = [p.name for p in party if p.dead]
    res["discovered_round"] = ctx["discovered_round"]
    res["result"] = "loss" if r2 < 0 else ("win" if r2 != 99 else "stalemate")
    return res

# --------------------------------------------------------------------- scoring
def score(res):
    """Returns (category, threat, triumph, drama)."""
    downs = sum(STATS["downs"].values())
    low = TRACK["low"]
    threat = min(10, (4 if low < 0.3 else 2 if low < 0.5 else 0) + 1.5 * min(3, downs)
                 + (1.5 if TRACK["brink"] else 0))
    triumph = min(10, min(3, TRACK["big"]) + 1.5 * min(2, TRACK["rescues"])
                  + (2 if res.get("discovered_round") and TRACK["holy_hit"] else 0)
                  + (2 if TRACK["comeback"] else 0))
    if res["result"] != "win":
        return "Loss", threat, triumph, 0.0
    pc_low = min(TRACK["pc_low"].values()) if TRACK["pc_low"] else 1.0
    if res["deaths"]: cat = "Pyrrhic"
    elif pc_low <= 0.25: cat = "Tense"
    elif pc_low < 0.5: cat = "Comfortable"
    else: cat = "Cruise"
    return cat, threat, triumph, math.sqrt(threat * triumph)

def run_many(n):
    global STATS, TRACK
    rows = []; agg = {"downs": Counter(), "deaths": Counter(), "boss_dmg": Counter(),
                      "pc_downs": Counter(), "pc_deaths": Counter()}
    for _ in range(n):
        STATS = new_stats(); TRACK = new_track()
        res = fight()
        cat, th, tr, dr = score(res)
        res.update(cat=cat, threat=th, triumph=tr, drama=dr,
                   dropped=bool(sum(STATS["downs"].values())))
        rows.append(res)
        for k in ("downs", "deaths", "boss_dmg"): agg[k].update(STATS[k])
        for name in res["deaths"]: agg["pc_deaths"][name] += 1
    return rows, agg

def summarize(rows):
    n = len(rows)
    cats = Counter(r["cat"] for r in rows)
    wins = [r for r in rows if r["result"] == "win"]
    great = sum(1 for r in wins if not r["deaths"] and r["threat"] >= 4 and r["triumph"] >= 4)
    flat = sum(1 for r in wins if r["threat"] < 2 or r["triumph"] < 2)
    p1 = [r["p1"] for r in rows if r["p1"] > 0]
    p2 = [r["p2"] for r in rows if r["p2"] and 0 < r["p2"] < 99]
    return dict(n=n, cats=cats, win=len(wins) / n, drama=st.mean(r["drama"] for r in rows),
                threat=st.mean(r["threat"] for r in rows), triumph=st.mean(r["triumph"] for r in rows),
                great=great / n, flat=flat / n, dropped=sum(r["dropped"] for r in rows) / n,
                p1=st.mean(p1) if p1 else float("nan"), p2=st.mean(p2) if p2 else float("nan"))

def report(rows, agg, title):
    s = summarize(rows); n = s["n"]; c = s["cats"]
    print(f"\n{title}  ({n:,} fights)")
    print(f"  Win {s['win']:.1%}   Loss {c['Loss'] / n:.1%}   Someone drops {s['dropped']:.0%}")
    print(f"  Outcomes: Cruise {c['Cruise'] / n:.0%} | Comfortable {c['Comfortable'] / n:.0%} | "
          f"Tense {c['Tense'] / n:.0%} | Pyrrhic {c['Pyrrhic'] / n:.0%} | Loss {c['Loss'] / n:.0%}")
    print(f"  Drama {s['drama']:.1f} (threat {s['threat']:.1f}, triumph {s['triumph']:.1f})   "
          f"Great finale {s['great']:.0%}   Flat {s['flat']:.0%}")
    print(f"  Rounds: Luna {s['p1']:.1f}, Avatar {s['p2']:.1f}")
    print("  PC deaths per fight: " + ", ".join(f"{k} {agg['pc_deaths'][k] / n:.1%}" for k in ("Kas", "Tavish", "Arabesque")))
    print("  What knocks PCs down (per fight): " + ", ".join(f"{k} {v / n:.2f}" for k, v in agg["downs"].most_common(5)))
    tot = sum(agg["boss_dmg"].values()) or 1
    share = Counter()
    for k, v in agg["boss_dmg"].items(): share[k.split(" ")[0]] += v
    print("  Damage to bosses: " + ", ".join(f"{k} {v / tot:.0%}" for k, v in share.most_common()))

def compare(n):
    print(f"\n{'preset':16s} {'win':>5s} {'loss':>5s} {'cruise':>7s} {'tense':>6s} {'drama':>6s} {'great':>6s} {'flat':>5s} {'rounds':>9s}")
    for name in PRESETS:
        configure(name, {})
        rows, _ = run_many(n); s = summarize(rows); c = s["cats"]
        print(f"{name:16s} {s['win']:5.0%} {c['Loss'] / n:5.0%} {c['Cruise'] / n:7.0%} {c['Tense'] / n:6.0%} "
              f"{s['drama']:6.1f} {s['great']:6.0%} {s['flat']:5.0%} {s['p1']:4.1f}+{s['p2']:.1f}")

def configure(preset, overrides):
    CFG.clear(); CFG.update(DEFAULTS); CFG.update(PRESETS[preset]); CFG.update(overrides)
    if CFG["level"] not in PARTY:
        raise SystemExit(f"No PARTY table for level {CFG['level']}")

def parse_sets(items):
    out = {}
    for item in items or []:
        key, _, raw = item.partition("=")
        if key not in DEFAULTS: raise SystemExit(f"Unknown knob '{key}'. Use --list to see them.")
        try: out[key] = ast.literal_eval(raw)
        except (ValueError, SyntaxError): out[key] = raw      # bare strings like adaptive
    return out

def main():
    global STATS, TRACK, LOGGING
    ap = argparse.ArgumentParser(description="Simulate the H.A.W.K. finale.")
    ap.add_argument("-n", type=int, default=5000, help="number of fights (default 5000)")
    ap.add_argument("--preset", default="recommended", choices=PRESETS)
    ap.add_argument("--set", action="append", metavar="KNOB=VALUE", help="override a knob; repeatable")
    ap.add_argument("--level", type=int, help="party level table to use (14 or 15)")
    ap.add_argument("--seed", type=int, help="random seed, for repeatable runs")
    ap.add_argument("--log", action="store_true", help="print one fight turn by turn")
    ap.add_argument("--compare", action="store_true", help="run every preset side by side")
    ap.add_argument("--list", action="store_true", help="list presets and knobs")
    a = ap.parse_args()
    if a.list:
        print("Presets:"); [print(f"  {k:16s} {v}") for k, v in PRESETS.items()]
        print("\nKnobs (defaults):"); [print(f"  {k:16s} {v!r}") for k, v in DEFAULTS.items()]
        return
    if a.seed is not None: R.seed(a.seed)
    overrides = parse_sets(a.set)
    if a.level: overrides["level"] = a.level
    if a.compare:
        compare(a.n); return
    configure(a.preset, overrides)
    if a.log:
        LOGGING = True; STATS = new_stats(); TRACK = new_track()
        res = fight(); cat, th, tr, dr = score(res)
        print("\n".join(LOG))
        print(f"\nResult: {res['result']} ({cat}) | rounds {res['p1']}+{res['p2']} | deaths {res['deaths'] or 'none'} | "
              f"threat {th:.1f}, triumph {tr:.1f}, drama {dr:.1f}")
        return
    rows, agg = run_many(a.n)
    shown = {k: v for k, v in overrides.items() if k != "level"}
    extra = f", {', '.join(f'{k}={v}' for k, v in shown.items())}" if shown else ""
    report(rows, agg, f"Preset '{a.preset}' (level {CFG['level']}{extra})")

if __name__ == "__main__":
    main()