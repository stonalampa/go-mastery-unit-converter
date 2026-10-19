# Shared CI workflow pinned to v1

CI for every go-mastery repo comes from one reusable workflow in the public `stonalampa/.github` repository, called with `uses: stonalampa/.github/.github/workflows/go-ci.yml@v1`, instead of each repo carrying its own copy. Callers pin a version tag rather than `@main` so a change to the shared workflow cannot break every repo at once; a repo moves to a new major tag deliberately. The required status check is therefore named `ci / check`, and branch protection refers to that name.
