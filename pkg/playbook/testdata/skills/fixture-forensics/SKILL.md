---
name: fixture-forensics
description: Test fixture for procedure loader
---

## When to use

- Disk imaging for incident response
- Evidence preservation before analysis

## Prerequisites

- Write-blocker attached

## Workflow

### Step 1: Network scan

Run `nmap` against the evidence network segment.

### Step 2: Create image

Use dd to capture the disk (manual steps).

## Tools & systems

- nmap
- dd

## Scenarios

- Ransomware disk capture
