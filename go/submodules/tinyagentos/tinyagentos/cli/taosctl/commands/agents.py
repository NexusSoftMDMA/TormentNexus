"""taosctl agents -- inspect and manage agents.

Reference noun: shows the pattern every other noun module follows (a NOUN, a
register() that wires verb subparsers, and small handlers that call the client
and return data for the framework to render).
"""
from __future__ import annotations

from urllib.parse import quote

NOUN = "agents"


def register(subparsers) -> None:
    p = subparsers.add_parser(NOUN, help="Inspect and manage agents")
    verbs = p.add_subparsers(dest="verb", required=True, metavar="<verb>")

    lp = verbs.add_parser("list", help="List all agents")
    lp.set_defaults(func=_list)

    gp = verbs.add_parser("get", help="Get one agent by name")
    gp.add_argument("name", help="Agent name")
    gp.set_defaults(func=_get)

    ap = verbs.add_parser("archived", help="List archived agents")
    ap.set_defaults(func=_archived)


def _list(args, client):
    return client.get("/api/agents")


def _get(args, client):
    return client.get(f"/api/agents/{quote(args.name, safe='')}")


def _archived(args, client):
    return client.get("/api/agents/archived")
