"""Prepare before the hour, then release a complete artifact no earlier than :00.

The release target defines the signed feed's validity window, not the time
weather readings were collected. The generator keeps actual collection times.
GitHub Pages activation/CDN latency remains outside this gate's control.
"""
import argparse
import datetime as dt
import os
from pathlib import Path
import time

HOUR = 3600
MAX_EARLY = 15 * 60
MAX_LATE = HOUR


def publication_target(now, requested=""):
    now = int(now)
    if requested:
        if not requested.isascii() or not requested.isdecimal():
            raise ValueError("publish_at must be an integer UTC Unix timestamp")
        target = int(requested)
    else:
        target = now // HOUR * HOUR
        if now % HOUR >= 50 * 60:
            target += HOUR
    if target < 946684800 or target % HOUR:
        raise ValueError("publish_at must be a UTC hour boundary after 2000")
    if target - now > MAX_EARLY:
        raise ValueError("Publication target is too far in the future")
    # Repair/manual jobs can publish within the current hour, without waiting
    # for another hour or claiming their collection time was the hour boundary.
    if requested and now - target >= MAX_LATE:
        raise ValueError("Requested publication cycle is too old; do not overwrite newer data")
    return target


def wait_for_release(target, clock=time.time, sleep=time.sleep):
    if target % HOUR:
        raise ValueError("Release target must be an hour boundary")
    while True:
        now = clock()
        if now >= target + HOUR:
            raise ValueError("Prepared feed has expired; refuse deployment")
        remaining = target - now
        if remaining <= 0:
            return now - target
        sleep(min(remaining, 30))


def utc_label(value):
    return dt.datetime.fromtimestamp(value, dt.timezone.utc).isoformat()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=("plan", "wait"))
    args = parser.parse_args()
    if args.mode == "plan":
        target = publication_target(time.time(), os.environ.get("REQUESTED_PUBLISH_AT", ""))
        output = os.environ.get("GITHUB_OUTPUT")
        if not output:
            raise ValueError("Missing GITHUB_OUTPUT")
        with Path(output).open("a", encoding="utf-8") as handle:
            handle.write(f"publish_at={target}\n")
        print(f"Preparing forecast valid from {utc_label(target)} to {utc_label(target + HOUR)}")
    else:
        requested = os.environ.get("PUBLISH_AT", "")
        if not requested.isascii() or not requested.isdecimal():
            raise ValueError("Missing/invalid PUBLISH_AT")
        target = int(requested)
        print(f"Holding prepared artifact until {utc_label(target)}", flush=True)
        late = wait_for_release(target)
        print(f"Release gate opened; {late:.3f} seconds after target. Pages activation still pending.", flush=True)


if __name__ == "__main__":
    main()
