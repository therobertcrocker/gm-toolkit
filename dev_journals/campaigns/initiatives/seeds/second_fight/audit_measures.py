"""Audit every drama measure against win chance.

Win chance (the party's chance of winning, estimated by playing the fight out from each round end)
is the yardstick: a close call should mean the party was really in danger, a big moment should move
the odds, and Dramatic fights should be more eventful than Flat ones.
"""
import sys, statistics as st, json
import hawk_finale_sim as H
import ythrel_sim as Y

N_ROLL = int(sys.argv[1]) if len(sys.argv) > 1 else 30
N_FIGHTS = int(sys.argv[2]) if len(sys.argv) > 2 else 1000

def ythrel_rows():
    Y.setcfg({}); Y.R.seed(11); Y.WINCHANCE_N = N_ROLL
    rows, _ = Y.run_many(N_FIGHTS); Y.WINCHANCE_N = 0
    return rows

def luna_rows():
    H.configure("recommended", {}); H.R.seed(11); H.WINCHANCE_N = N_ROLL
    rows, _ = H.run_many(N_FIGHTS); H.WINCHANCE_N = 0
    return rows

def audit(name, rows):
    ms = [r["m"] for r in rows]
    p0 = sum(m["win"] for m in ms) / len(ms)           # win chance before the fight starts
    wins = [m for m in ms if m["win"] and m["wc"]]
    out = {"name": name, "fights": len(ms), "wins": len(wins), "p0": p0}
    for m in wins:
        curve = [p0] + m["wc"] + [1.0]
        steps = [b - a for a, b in zip(curve, curve[1:])]
        m["min_wc"] = min(m["wc"])
        m["excite"] = sum(abs(x) for x in steps)
        m["up_swings"] = sum(1 for x in steps[:-1] if x >= 0.15)     # the party's odds jump 15+ points in a round
        m["big_swing"] = max(steps[:-1] or [0]) >= 0.15
    def rate(sub, f): return sum(f(m) for m in sub) / len(sub) if sub else float("nan")
    danger = lambda m: m["min_wc"] <= 0.5
    worry = lambda m: m["min_wc"] <= 0.75
    close = [m for m in wins if m["close"]]; notclose = [m for m in wins if not m["close"]]
    out["close"] = dict(rate=rate(wins, lambda m: m["close"]),
                        danger_if_close=rate(close, danger), danger_if_not=rate(notclose, danger),
                        worry_if_close=rate(close, worry), worry_if_not=rate(notclose, worry),
                        real_danger=rate(wins, danger), real_worry=rate(wins, worry),
                        recall=rate([m for m in wins if danger(m)], lambda m: m["close"]))
    big = [m for m in wins if m["big"]]; notbig = [m for m in wins if not m["big"]]
    out["big"] = dict(rate=rate(wins, lambda m: m["big"]),
                      swing_if_big=rate(big, lambda m: m["big_swing"]), swing_if_not=rate(notbig, lambda m: m["big_swing"]),
                      real_swing=rate(wins, lambda m: m["big_swing"]),
                      corr=_corr([m["moments"] for m in wins], [m["up_swings"] for m in wins]))
    out["lead"] = dict(doubt=rate(wins, lambda m: m["lead"]), hp_proxy=rate(wins, lambda m: m["lead_hp"]))
    tiers = {"Flat (0/3)": lambda m: m["k"] == 0, "One (1/3)": lambda m: m["k"] == 1,
             "Leans (2/3)": lambda m: m["k"] == 2, "Dramatic (3/3)": lambda m: m["k"] == 3}
    out["tiers"] = {t: (rate(wins, f), st.mean([m["excite"] for m in wins if f(m)] or [float("nan")]))
                    for t, f in tiers.items()}
    out["excite_all"] = st.mean(m["excite"] for m in wins)
    return out

def _corr(a, b):
    if len(a) < 3 or st.pstdev(a) == 0 or st.pstdev(b) == 0: return float("nan")
    ma, mb = st.mean(a), st.mean(b)
    return sum((x - ma) * (y - mb) for x, y in zip(a, b)) / (len(a) * st.pstdev(a) * st.pstdev(b))

def show(o):
    c, b, l = o["close"], o["big"], o["lead"]
    print(f"\n=== {o['name']}: {o['fights']} fights, {o['wins']} wins, {N_ROLL} playouts per round end; win chance at the start {o['p0']:.0%}")
    print(f"Close call (HP to 25% or 2+ KOs): passes {c['rate']:.0%} of wins")
    print(f"   real danger (win chance fell to 50% or less): {c['real_danger']:.0%} of wins | when close: {c['danger_if_close']:.0%}, when not: {c['danger_if_not']:.0%} | catches {c['recall']:.0%} of real danger")
    print(f"   real worry (win chance fell to 75% or less):  {c['real_worry']:.0%} of wins | when close: {c['worry_if_close']:.0%}, when not: {c['worry_if_not']:.0%}")
    print(f"Big moments (1 per 2 rounds): passes {b['rate']:.0%} of wins")
    print(f"   real swing (odds jump 15+ points in a round): {b['real_swing']:.0%} of wins | when big: {b['swing_if_big']:.0%}, when not: {b['swing_if_not']:.0%} | moments vs swings correlation {b['corr']:.2f}")
    print(f"In doubt late (win chance 25-75% in 2nd half): {l['doubt']:.0%} of wins   (HP-race proxy said {l['hp_proxy']:.0%})")
    print(f"Tiers vs how eventful the fight was (excitement = total movement in win chance; average {o['excite_all']:.2f}):")
    for t, (r, e) in o["tiers"].items():
        print(f"   {t:15s} {r:4.0%} of wins   excitement {e:.2f}")

if __name__ == "__main__":
    results = [audit("Ythrel (chart build)", ythrel_rows()), audit("Luna (recommended)", luna_rows())]
    for o in results: show(o)
    json.dump(results, open("audit_results.json", "w"), indent=1, default=str)
