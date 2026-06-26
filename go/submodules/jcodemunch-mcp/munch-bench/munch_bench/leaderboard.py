"""Leaderboard generator — produces static HTML with tables and charts."""

from __future__ import annotations

import os
from typing import Optional

from jinja2 import Template

from .evaluate import BenchmarkRun


LEADERBOARD_TEMPLATE = Template("""\
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>munch-bench Leaderboard</title>
    <style>
        :root {
            --bg: #0d1117; --surface: #161b22; --border: #30363d;
            --text: #e6edf3; --text-muted: #8b949e;
            --accent: #58a6ff; --green: #3fb950; --orange: #d29922; --red: #f85149;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
            background: var(--bg); color: var(--text); line-height: 1.5;
            max-width: 1200px; margin: 0 auto; padding: 2rem;
        }
        h1 { font-size: 2rem; margin-bottom: 0.25rem; }
        .subtitle { color: var(--text-muted); margin-bottom: 2rem; font-size: 0.95rem; }
        .subtitle a { color: var(--accent); text-decoration: none; }

        table { width: 100%; border-collapse: collapse; margin-bottom: 2rem; }
        th, td { padding: 0.75rem 1rem; text-align: left; border-bottom: 1px solid var(--border); }
        th { background: var(--surface); color: var(--text-muted); font-weight: 600; font-size: 0.85rem;
             text-transform: uppercase; letter-spacing: 0.05em; }
        tr:hover { background: var(--surface); }
        td.num { text-align: right; font-variant-numeric: tabular-nums; }

        .rank { font-weight: 700; color: var(--accent); }
        .bar-cell { position: relative; }
        .bar { height: 6px; border-radius: 3px; background: var(--accent); min-width: 2px; }
        .bar.green { background: var(--green); }
        .bar.orange { background: var(--orange); }

        .charts { display: grid; grid-template-columns: 1fr 1fr; gap: 2rem; margin-bottom: 2rem; }
        .chart-box { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 1.5rem; }
        .chart-box h3 { font-size: 0.95rem; color: var(--text-muted); margin-bottom: 1rem; }
        canvas { width: 100% !important; height: 250px !important; }

        .footer { color: var(--text-muted); font-size: 0.8rem; text-align: center; margin-top: 3rem;
                  padding-top: 1.5rem; border-top: 1px solid var(--border); }
        .footer a { color: var(--accent); text-decoration: none; }

        @media (max-width: 768px) {
            .charts { grid-template-columns: 1fr; }
            body { padding: 1rem; }
            th, td { padding: 0.5rem; font-size: 0.9rem; }
        }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>
</head>
<body>
    <h1>munch-bench Leaderboard</h1>
    <p class="subtitle">
        Retrieval + Inference benchmark for LLM-powered codebase Q&A &mdash;
        Powered by <a href="https://github.com/jgravelle/jcodemunch-mcp">jCodeMunch</a> +
        <a href="https://groq.com">Groq</a>
    </p>

    <table>
        <thead>
            <tr>
                <th>#</th>
                <th>Provider / Model</th>
                <th>Judge Score</th>
                <th></th>
                <th>P@5</th>
                <th>Recall</th>
                <th>Exact Match</th>
                <th>Avg Time</th>
                <th>Cost</th>
                <th>Questions</th>
            </tr>
        </thead>
        <tbody>
            {% for entry in entries %}
            <tr>
                <td class="rank">{{ entry.rank }}</td>
                <td><strong>{{ entry.provider }}</strong> / {{ entry.model }}</td>
                <td class="num">{{ "%.2f"|format(entry.judge_score) }}</td>
                <td class="bar-cell">
                    <div class="bar{% if entry.judge_score >= 0.7 %} green{% elif entry.judge_score >= 0.4 %} orange{% endif %}"
                         style="width: {{ (entry.judge_score * 100)|int }}%"></div>
                </td>
                <td class="num">{{ "%.2f"|format(entry.p_at_5) }}</td>
                <td class="num">{{ "%.2f"|format(entry.recall) }}</td>
                <td class="num">{{ "%.1f%%"|format(entry.exact_match * 100) }}</td>
                <td class="num">{{ "%.2fs"|format(entry.avg_time) }}</td>
                <td class="num">${{ "%.4f"|format(entry.cost) }}</td>
                <td class="num">{{ entry.n_questions }}</td>
            </tr>
            {% endfor %}
        </tbody>
    </table>

    <div class="charts">
        <div class="chart-box">
            <h3>LLM Judge Score by Model</h3>
            <canvas id="chartJudge"></canvas>
        </div>
        <div class="chart-box">
            <h3>Cost vs Accuracy</h3>
            <canvas id="chartCost"></canvas>
        </div>
    </div>

    <script>
        const labels = {{ labels | safe }};
        const judgeScores = {{ judge_scores | safe }};
        const costs = {{ costs | safe }};

        new Chart(document.getElementById('chartJudge'), {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Judge Score',
                    data: judgeScores,
                    backgroundColor: judgeScores.map(s => s >= 0.7 ? '#3fb950' : s >= 0.4 ? '#d29922' : '#f85149'),
                    borderRadius: 4,
                }]
            },
            options: {
                responsive: true,
                scales: { y: { beginAtZero: true, max: 1.0, ticks: { color: '#8b949e' }, grid: { color: '#30363d' } },
                          x: { ticks: { color: '#8b949e' }, grid: { display: false } } },
                plugins: { legend: { display: false } }
            }
        });

        new Chart(document.getElementById('chartCost'), {
            type: 'scatter',
            data: {
                datasets: [{
                    label: 'Models',
                    data: labels.map((l, i) => ({ x: costs[i], y: judgeScores[i], label: l })),
                    backgroundColor: '#58a6ff',
                    pointRadius: 8,
                }]
            },
            options: {
                responsive: true,
                scales: {
                    x: { title: { display: true, text: 'Cost (USD)', color: '#8b949e' }, ticks: { color: '#8b949e' }, grid: { color: '#30363d' } },
                    y: { title: { display: true, text: 'Judge Score', color: '#8b949e' }, beginAtZero: true, max: 1.0, ticks: { color: '#8b949e' }, grid: { color: '#30363d' } }
                },
                plugins: {
                    tooltip: {
                        callbacks: { label: ctx => ctx.raw.label + ': $' + ctx.raw.x.toFixed(4) + ', score=' + ctx.raw.y.toFixed(2) }
                    },
                    legend: { display: false }
                }
            }
        });
    </script>

    <div class="footer">
        Generated by <a href="https://github.com/jgravelle/munch-bench">munch-bench</a> &mdash;
        Powered by <a href="https://github.com/jgravelle/jcodemunch-mcp">jCodeMunch</a> + <a href="https://groq.com">Groq</a>
    </div>
</body>
</html>
""")


def generate_leaderboard(runs: list[BenchmarkRun], output_path: str = "leaderboard.html") -> None:
    """Generate a static HTML leaderboard from benchmark runs."""
    # Sort by judge score descending
    sorted_runs = sorted(runs, key=lambda r: r.avg_llm_judge_score, reverse=True)

    entries = []
    labels = []
    judge_scores = []
    costs = []

    for i, run in enumerate(sorted_runs):
        label = f"{run.provider}/{run.model}"
        entries.append({
            "rank": i + 1,
            "provider": run.provider,
            "model": run.model,
            "judge_score": run.avg_llm_judge_score,
            "p_at_5": run.avg_retrieval_precision_at_5,
            "recall": run.avg_retrieval_recall,
            "exact_match": run.exact_match_rate,
            "avg_time": run.avg_wall_time_s,
            "cost": run.total_cost_usd,
            "n_questions": len(run.results),
        })
        labels.append(label)
        judge_scores.append(round(run.avg_llm_judge_score, 4))
        costs.append(round(run.total_cost_usd, 6))

    import json
    html = LEADERBOARD_TEMPLATE.render(
        entries=entries,
        labels=json.dumps(labels),
        judge_scores=json.dumps(judge_scores),
        costs=json.dumps(costs),
    )

    os.makedirs(os.path.dirname(output_path) or ".", exist_ok=True)
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(html)
