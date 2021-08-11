# declcfg

`declcfg` is a small toolkit that implements several useful idempotent functions for DC-based packages

## Install

### Releases

Binary downloads are available on the [releases](https://github.com/joelanford/declcfg/releases) page.

### From source
```bash
( 
  TMPDIR=$(mktemp -d)
  git clone https://github.com/joelanford/declcfg $TMPDIR
  make -C $TMPDIR install
  rm -rf $TMPDIR
)
```

## Examples

### `inherit-channels`

The `inherit-channels` subcommand ensures that all bundles in a given replaces chain are present in all of the channels in which bundles on that chain are members.

For example, consider an index with:
- `v0.1.0` (channels: `alpha`)
- `v0.1.1` (channels: `beta`, replaces: `v0.1.0`)
- `v0.1.2` (channels: `stable`, replaces: `v0.1.1`)

Running `inherit-channels` against this index will update the index to:
- `v0.1.0` (channels: `alpha`, `beta`, `stable`)
- `v0.1.1` (channels: `beta`, `stable`, replaces: `v0.1.0`)
- `v0.1.2` (channels: `stable`, replaces: `v0.1.1`)

See [examples/inherit-channels/build.sh](examples/inherit-channels/build.sh)

Without channel inheritance, users who have `v0.1.0` installed and want to upgrade to `v0.3.0` must:
1. Subscribe to `beta` and wait for `v0.2.0` to be installed by OLM.
1. Subscribe to `stable` and wait for `v0.3.0` to be installed by OLM.

With channel inheritance, users who have `v0.1.0` installed and want to upgrade to `v0.3.0` can update their subscription directly to `stable`. Since `v0.1.0` and `v0.1.1` inherited the `stable` channel from `v0.1.2`'s replaces chain, OLM will traverse upgrades to `v0.3.0`.

### `semver`

The `semver` subcommand re-orders channels based on the semver versions of the bundles in the channel. When `--skip-patch` is not enabled, `semver` simply creates an ordered replaces chain based on a sort of the bundles' semver versions. When `--skip-patch` is enabled, the replaces chain is contains just the max z-stream versions from each y-stream in a channel. All other bundles are skipped.

Consider bundles with the following versions:
- `v0.1.0`
- `v0.1.1`
- `v0.1.2`
- `v0.1.3`
- `v0.2.0`
- `v0.2.1`
- `v0.2.2`
- `v0.2.3`
- `v0.3.0`
- `v0.3.1`
- `v0.3.2`
- `v0.3.3`

When `--skip-patch` is disabled, these are ordered as is, regardless of their existing `replaces` values.
When `--skip-patch` is enabled, a shorter replaces chain is created, with intermediate bundles being skipped.

As shown below, versions with `*` are in the replacement chain, and versions with `X` are skipped by the highest patch
version in that particular z-stream.
```
* v0.1.0
|  X v0.1.1
|  X v0.1.2
* v0.1.3
|  X v0.2.0
|  X v0.2.1
|  X v0.2.2
* v0.2.3
|  X v0.3.0
|  X v0.3.1
|  X v0.3.2
* v0.3.3
```

See [examples/semver/build.sh](examples/semver/build.sh)

### `inline-bundles`

The `inline-bundles` command updates the `olm.bundle.object` properties by pulling bundle images and inlining their manifests. This command can also idempotently prune `olm.bundle.object` properties from non-channel-head bundles.

For example, to make sure all channel heads (and only channel heads) for a particular package have `olm.bundle.object` properties present, just run the following:
```
$ declcfg inline-bundles my-index my-package --prune
```
