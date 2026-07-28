#!/usr/bin/env python3
from __future__ import annotations

import argparse
from pathlib import Path
import yaml

ROOT = Path(__file__).resolve().parents[2]
CATALOG = ROOT / "engage/serve/catalog/tools.yaml"
OUT = ROOT / "scripts/ops/engage-tools-sources.yaml"


CURATED: dict[str, dict[str, object]] = {
    "httpx": {
        "binary": "httpx",
        "kali_tool_page": "https://kali.org/tools/httpx-toolkit",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/httpx-toolkit/",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/httpx-toolkit",
        "upstream_repo": "https://github.com/projectdiscovery/httpx",
        "preferred_install_methods": [
            "apt:httpx-toolkit",
            "go:github.com/projectdiscovery/httpx/cmd/httpx@latest",
        ],
    },
    "nuclei": {
        "binary": "nuclei",
        "kali_tool_page": "https://kali.org/tools/nuclei",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/nuclei",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/nuclei",
        "upstream_repo": "https://github.com/projectdiscovery/nuclei",
        "preferred_install_methods": [
            "apt:nuclei",
            "go:github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest",
        ],
    },
    "subfinder": {
        "binary": "subfinder",
        "kali_tool_page": "https://kali.org/tools/subfinder",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/subfinder",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/subfinder",
        "upstream_repo": "https://github.com/projectdiscovery/subfinder",
        "preferred_install_methods": [
            "apt:subfinder",
            "go:github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest",
        ],
    },
    "amass": {
        "binary": "amass",
        "kali_tool_page": "https://kali.org/tools/amass",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/amass",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/amass",
        "upstream_repo": "https://github.com/owasp-amass/amass",
        "preferred_install_methods": [
            "apt:amass",
            "go:github.com/owasp-amass/amass/v4/...@master",
        ],
    },
    "gobuster": {
        "binary": "gobuster",
        "kali_tool_page": "https://kali.org/tools/gobuster",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/gobuster",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/gobuster",
        "upstream_repo": "https://github.com/OJ/gobuster",
        "preferred_install_methods": [
            "apt:gobuster",
            "go:github.com/OJ/gobuster/v3@latest",
        ],
    },
    "feroxbuster": {
        "binary": "feroxbuster",
        "kali_tool_page": "https://kali.org/tools/feroxbuster",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/feroxbuster",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/feroxbuster",
        "upstream_repo": "https://github.com/epi052/feroxbuster",
        "preferred_install_methods": [
            "apt:feroxbuster",
            "cargo:feroxbuster",
        ],
    },
    "ffuf": {
        "binary": "ffuf",
        "kali_tool_page": "https://kali.org/tools/ffuf",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/ffuf",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/ffuf",
        "upstream_repo": "https://github.com/ffuf/ffuf",
        "preferred_install_methods": [
            "apt:ffuf",
            "go:github.com/ffuf/ffuf/v2@latest",
        ],
    },
    "sqlmap": {
        "binary": "sqlmap",
        "kali_tool_page": "https://kali.org/tools/sqlmap",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/sqlmap",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/sqlmap",
        "upstream_repo": "https://github.com/sqlmapproject/sqlmap",
        "preferred_install_methods": ["apt:sqlmap"],
    },
    "nikto": {
        "binary": "nikto",
        "kali_tool_page": "https://kali.org/tools/nikto",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/nikto",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/nikto",
        "upstream_repo": "https://github.com/sullo/nikto",
        "preferred_install_methods": ["apt:nikto"],
    },
    "masscan": {
        "binary": "masscan",
        "kali_tool_page": "https://kali.org/tools/masscan",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/masscan",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/masscan",
        "upstream_repo": "https://github.com/robertdavidgraham/masscan",
        "preferred_install_methods": ["apt:masscan"],
    },
    "httpx": {
        "binary": "httpx",
        "kali_tool_page": "https://kali.org/tools/httpx-toolkit",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/httpx-toolkit/",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/httpx-toolkit",
        "upstream_repo": "https://github.com/projectdiscovery/httpx",
        "preferred_install_methods": [
            "apt:httpx-toolkit",
            "go:github.com/projectdiscovery/httpx/cmd/httpx@latest",
        ],
    },
    "nuclei_scan": {
        "binary": "nuclei",
        "kali_tool_page": "https://kali.org/tools/nuclei",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/nuclei",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/nuclei",
        "upstream_repo": "https://github.com/projectdiscovery/nuclei",
        "preferred_install_methods": [
            "apt:nuclei",
            "go:github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest",
        ],
    },
    "subdomain_discovery": {
        "binary": "subfinder",
        "kali_tool_page": "https://kali.org/tools/subfinder",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/subfinder",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/subfinder",
        "upstream_repo": "https://github.com/projectdiscovery/subfinder",
        "preferred_install_methods": [
            "apt:subfinder",
            "go:github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest",
        ],
    },
    "amass_intel": {
        "binary": "amass",
        "kali_tool_page": "https://kali.org/tools/amass",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/amass",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/amass",
        "upstream_repo": "https://github.com/owasp-amass/amass",
        "preferred_install_methods": [
            "apt:amass",
            "go:github.com/owasp-amass/amass/v4/...@master",
        ],
    },
    "feroxbuster_scan": {
        "binary": "feroxbuster",
        "kali_tool_page": "https://kali.org/tools/feroxbuster",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/feroxbuster",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/feroxbuster",
        "upstream_repo": "https://github.com/epi052/feroxbuster",
        "preferred_install_methods": [
            "apt:feroxbuster",
            "cargo:feroxbuster",
        ],
    },
    "ffuf_scan": {
        "binary": "ffuf",
        "kali_tool_page": "https://kali.org/tools/ffuf",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/ffuf",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/ffuf",
        "upstream_repo": "https://github.com/ffuf/ffuf",
        "preferred_install_methods": [
            "apt:ffuf",
            "go:github.com/ffuf/ffuf/v2@latest",
        ],
    },
    "gobuster_scan": {
        "binary": "gobuster",
        "kali_tool_page": "https://kali.org/tools/gobuster",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/gobuster",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/gobuster",
        "upstream_repo": "https://github.com/OJ/gobuster",
        "preferred_install_methods": [
            "apt:gobuster",
            "go:github.com/OJ/gobuster/v3@latest",
        ],
    },
    "sqlmap_scan": {
        "binary": "sqlmap",
        "kali_tool_page": "https://kali.org/tools/sqlmap",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/sqlmap",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/sqlmap",
        "upstream_repo": "https://github.com/sqlmapproject/sqlmap",
        "preferred_install_methods": ["apt:sqlmap"],
    },
    "nikto_scan": {
        "binary": "nikto",
        "kali_tool_page": "https://kali.org/tools/nikto",
        "kali_pkg_tracker": "https://pkg.kali.org/pkg/nikto",
        "kali_packaging_repo": "https://gitlab.com/kalilinux/packages/nikto",
        "upstream_repo": "https://github.com/sullo/nikto",
        "preferred_install_methods": ["apt:nikto"],
    },
}


def load_catalog_tools() -> list[dict[str, object]]:
    data = yaml.safe_load(CATALOG.read_text(encoding="utf-8"))
    return list(data.get("tools", []))


def fallback_record(name: str, binary: str) -> dict[str, object]:
    # Default guessable references. Curated overrides replace these where known.
    slug = name.replace("_", "-")
    pkg_guess = binary if binary else slug
    return {
        "binary": binary,
        "kali_tool_page": f"https://kali.org/tools/{slug}",
        "kali_pkg_tracker": f"https://pkg.kali.org/pkg/{pkg_guess}",
        "kali_packaging_repo": f"https://gitlab.com/kalilinux/packages/{pkg_guess}",
        "upstream_repo": "",
        "preferred_install_methods": [],
    }


def generate() -> dict[str, object]:
    tools = load_catalog_tools()
    out_tools: dict[str, dict[str, object]] = {}
    for t in tools:
        name = str(t.get("name", "")).strip()
        binary = str(t.get("binary", "")).strip()
        if not name:
            continue
        rec = fallback_record(name, binary)
        if name in CURATED:
            rec.update(CURATED[name])
        elif binary in CURATED:
            rec.update(CURATED[binary])
        out_tools[name] = rec
        if binary and binary not in out_tools:
            # Alias by executable name to support preflight/installer flows that operate on binaries.
            alias = fallback_record(binary, binary)
            if binary in CURATED:
                alias.update(CURATED[binary])
            elif name in CURATED:
                alias.update(CURATED[name])
            out_tools[binary] = alias
    return {"schema_version": 2, "tools": out_tools}


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--check", action="store_true")
    args = ap.parse_args()
    payload = generate()
    rendered = yaml.safe_dump(payload, sort_keys=False, allow_unicode=False)
    if args.check:
        existing = OUT.read_text(encoding="utf-8") if OUT.exists() else ""
        if existing != rendered:
            print(f"stale: {OUT}")
            return 1
        print(f"ok: {OUT}")
        return 0
    OUT.write_text(rendered, encoding="utf-8")
    print(f"wrote {OUT} ({len(payload['tools'])} tools)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
