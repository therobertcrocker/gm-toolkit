"""Does the HP-race 'lead change' measure match the real odds?

For a sample of fights, at the end of every round we copy the whole fight state and play it
out N times to estimate the party's chance of winning from there. Then we compare:
  HP race:  the party's HP share is below the boss's at some round end in the second half
  Win odds: the party's win chance is below 50% at some round end in the second half
both only for fights the party went on to win.
"""
import copy, sys, statistics as st
import ythrel_sim as Y, hawk_finale_sim as H

ROLLOUTS = int(sys.argv[1]) if len(sys.argv) > 1 else 24
FIGHTS = int(sys.argv[2]) if len(sys.argv) > 2 else 500

def continue_fight(party, y, ctx, order, start):
    """The round loop of ythrel_sim.fight(), resumed at round `start`. Returns True on a win."""
    R, CFG = Y.R, Y.CFG
    for rnd in range(start, 16):
        ctx["round"] = rnd; ctx["named"] = set(); ctx["round_best"] = (None, 0); ctx["vuln_used"] = set()
        for c in order:
            if y.hp <= 0: return True
            if not any(p.up for p in party): return False
            if c is y:
                Y.ythrel_turn(y, party, ctx)
                if ctx["sigh_cd"]: ctx["sigh_cd"] -= 1
            else:
                if c.dead: continue
                if c.hp <= 0: H.dying_check(c); continue
                Y.pc_turn(c, party, ctx); Y.end_of_turn(c)
        if y.hp <= 0: return True
        if not any(p.up for p in party): return False
        ctx["hardest"] = ctx["round_best"][0]
        if not ctx["discovered"]:
            curve = CFG["disc_curve"] or {1: 0.0, 2: 0.25, 3: 0.5, 4: 0.6, 5: 1.0}
            pr = curve.get(rnd, 1.0)
            if rnd < 3 and not ctx["flaked"]: pr = 0
            if R.random() < pr: ctx["discovered"] = True; ctx["discovered_round"] = rnd
    return False

def win_chance(state, start):
    wins = 0
    saved = (Y.STATS, H.STATS, H.TRACK, Y.CTX)
    for _ in range(ROLLOUTS):
        party, y, ctx, order, stats, track = copy.deepcopy(state)
        Y.STATS = H.STATS = stats; H.TRACK = track; Y.CTX = ctx
        wins += continue_fight(party, y, ctx, order, start)
    Y.STATS, H.STATS, H.TRACK, Y.CTX = saved
    return wins / ROLLOUTS

def one_fight():
    """ythrel_sim.fight(), with a win-chance estimate at the end of every round."""
    curve = []
    orig_round_end = H.round_end
    def hooked(party, frac):
        orig_round_end(party, frac)
        ctx = Y.CTX; y = ctx["ythrel"]
        if y.hp > 0 and any(p.up for p in party):
            order = ctx["_order"]
            state = (party, y, ctx, order, Y.STATS, H.TRACK)
            curve.append(win_chance(state, ctx["round"] + 1))
    H.round_end = hooked
    orig_shuffle = Y.R.shuffle
    def grab(seq):
        orig_shuffle(seq)
        if Y.CTX is not None and len(seq) == 4: Y.CTX["_order"] = seq
    Y.R.shuffle = grab
    try:
        res = Y.fight()
    finally:
        H.round_end = orig_round_end; Y.R.shuffle = orig_shuffle
    return res, curve, list(H.TRACK["diffs"])

def main():
    Y.setcfg({}); Y.R.seed(7)
    both = hp_only = wp_only = neither = 0
    doubt = 0; wins = 0; deltas = []
    for i in range(FIGHTS):
        Y.STATS = Y.new_stats(); H.STATS = Y.STATS; H.TRACK = H.new_track()
        res, curve, diffs = one_fight()
        if res["result"] != "win" or not diffs: continue
        wins += 1
        late_d = diffs[len(diffs) // 2:]
        late_c = curve[len(curve) // 2:] if curve else []
        hp = any(x < -0.02 for x in late_d)
        wp = any(c < 0.5 for c in late_c)
        doubt += any(0.25 <= c <= 0.75 for c in late_c)
        both += hp and wp; hp_only += hp and not wp; wp_only += wp and not hp; neither += not hp and not wp
        if curve: deltas.append(min(curve))
    n = wins
    print(f"{FIGHTS} fights, {wins} wins, {ROLLOUTS} rollouts per round end")
    print(f"  HP race says behind late:   {(both + hp_only) / n:.0%}")
    print(f"  Win odds say under 50% late: {(both + wp_only) / n:.0%}")
    print(f"  Agree (both or neither): {(both + neither) / n:.0%}   HP race only: {hp_only / n:.0%}   win odds only: {wp_only / n:.0%}")
    print(f"  Result in doubt late (win odds 25-75% in the second half): {doubt / n:.0%}")
    print(f"  Lowest win chance during a win, median: {st.median(deltas):.0%}")

if __name__ == "__main__":
    main()
