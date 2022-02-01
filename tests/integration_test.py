import os
import json
import subprocess
import pytest

from typing import Set, Optional, List

GH_INTEGRATION_TESTS_TOKEN = os.environ["GH_INTEGRATION_TESTS_TOKEN"]

def run_src_fingerprint(provider: str, args: Optional[List[str]] = []):
    subprocess.run(
        [
            "./src-fingerprint",
            "-p",
            provider,
            "--token",
            GH_INTEGRATION_TESTS_TOKEN,
            "-f",
            "jsonl",
            "-o",
            "fingerprints.jsonl"
        ] + args,
        check=True
    )

def load_jsonl(jsonl_path):
    with open(jsonl_path) as f:
        for line in f:
            yield json.loads(line)

def get_output_repos(output_path) -> Set[str]:
    return {x["repository_name"] for x in load_jsonl(output_path)}


def test_local_repository():
    run_src_fingerprint(provider="repository", args=["--object", "."])
    repos = get_output_repos("fingerprints.jsonl")
    os.remove("fingerprints.jsonl")
    assert len(repos) == 1


@pytest.mark.parametrize(
    "title, cmd_args, expected_output_repos", [
        (
            "Get all private repos accesible to user gg-src-fingerprint except archived ones",
            [], 
            {
                # Repos for user gg-src-fingerprint
                "main_private",
                # Repos for org gg-src-fingerprint-org
                "external_private",
                "gg_src_fingerprint_private",
            }
        ),
        (
            "Get all private repos accesible to user gg-src-fingerprint",
            ["--include-archived-repos"], 
            {
                # Repos for user gg-src-fingerprint
                "main_private_archive",
                "main_private",
                # Repos for org gg-src-fingerprint-org
                "gg_src_fingerprint_private",
                "gg_src_fingerprint_private_archive",
                "external_private",
            }
        ),
        (
            "Get all repos accesible to user gg-src-fingerprint except archived ones",
            ["--include-public-repos", "--include-forked-repos"], 
            {
                # Repos for user gg-src-fingerprint
                "main_private",
                "main_public",
                "src-fingerprint",
                # Repos for org gg-src-fingerprint-org
                "external_private",
                "gg_src_fingerprint_private",
                "gg_src_fingerprint_public",
            }
        ),
        (
            "Get all repos accesible to user gg-src-fingerprint except forked ones",
            ["--include-public-repos", "--include-archived-repos"], 
            {
                # Repos for user gg-src-fingerprint
                "main_archive",
                "main_private",
                "main_private_archive",
                "main_public",
                # Repos for org gg-src-fingerprint-org
                "external_private",
                "gg_src_fingerprint_private",
                "gg_src_fingerprint_private_archive",
                "gg_src_fingerprint_public",
                "gg_src_fingerprint_public_archive",
            }
        ),
                (
            "Get all repos accesible to user gg-src-fingerprint",
            ["--include-public-repos", "--include-archived-repos", "--include-forked-repos"], 
            {
                # Repos for user gg-src-fingerprint
                "main_archive",
                "main_private",
                "main_private_archive",
                "main_public",
                "src-fingerprint",
                # Repos for org gg-src-fingerprint-org
                "external_private",
                "gg_src_fingerprint_private",
                "gg_src_fingerprint_private_archive",
                "gg_src_fingerprint_public",
                "gg_src_fingerprint_public_archive",
            }
        )
    ]
)
def test_src_fingerprint_no_object_specified(title, cmd_args, expected_output_repos):
    run_src_fingerprint(provider="github", args=cmd_args)
    output_repos = get_output_repos("fingerprints.jsonl")
    os.remove("fingerprints.jsonl")
    assert output_repos == expected_output_repos


@pytest.mark.parametrize(
    "title, cmd_args, expected_output_repos", [
        (
            "Get all private repos except archived ones for org gg-src-fingerprint-org",
            [], 
            {
                "gg_src_fingerprint_private",
                "external_private"
            }
        ),
        (
            "Get all private repos for org gg-src-fingerprint-org",
            ["--include-archived-repos"], 
            {
                "gg_src_fingerprint_private",
                "gg_src_fingerprint_private_archive",
                "external_private",
            }
        ),
        (
            "Get all repos for org gg-src-fingerprint-org except archived ones",
            ["--include-public-repos", "--include-forked-repos"], 
            {
                "external_private",
                "gg_src_fingerprint_private",
                "gg_src_fingerprint_public",
                "src-fingerprint",
            }
        ),
        (
            "Get all repos for org gg-src-fingerprint-org except forked ones",
            ["--include-public-repos", "--include-archived-repos"], 
            {
                "external_private",
                "gg_src_fingerprint_private",
                "gg_src_fingerprint_private_archive",
                "gg_src_fingerprint_public",
                "gg_src_fingerprint_public_archive",
            }
        ),
                (
            "Get all repos for org gg-src-fingerprint-org",
            ["--include-public-repos", "--include-archived-repos", "--include-forked-repos"], 
            {
                "external_private",
                "gg_src_fingerprint_private",
                "gg_src_fingerprint_private_archive",
                "gg_src_fingerprint_public",
                "gg_src_fingerprint_public_archive",
                "src-fingerprint",
            }
        )
    ]
)
def test_src_fingerprint_on_org(title, cmd_args, expected_output_repos):
    run_src_fingerprint(provider="github", args=["--object", "gg-src-fingerprint-org"]+cmd_args)
    output_repos = get_output_repos("fingerprints.jsonl")
    os.remove("fingerprints.jsonl")
    assert output_repos == expected_output_repos