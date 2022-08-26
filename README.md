# Snapshoter
## V0.1
### Simple Tool to make file Snapshots

Snapshoter allows to take a snapshot of a source
directory at regular intervals. The cli arguments can be
used to manage source path, target path and time period.

Usage:
Snapshoter [flags]

Flags:

    --config string            Path to a YAML config file (default is $HOME/.cobra)
    -d, --destination string   Source directory
    -h, --help                 help for Snapshoter
    -m, --max_shots int        The maximum number of snapshots to be stored
    -p, --period int           Snapshot period in hoers
    -s, --source string        Source directory
