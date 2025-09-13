# Rclone Sync with Prometheus Metrics

This is an rclone sync wrapper that pushes metrics to prometheus
upon completion. This is a small piece of a much larger distributed
backup system. It exists to run on a fan-in point for restic backups
to replicate the encrypted repositories to Backblaze B2 as a form of
off-site backup.

This is a port of some old python scripts that used a patched version
of rclone that emitted a JSON statistics log that was then parsed and
pushed to prometheus. This approach should be less fragile and relies
only on the public API surface of rclone.

## Building

This can be built pretty simply by checking out the code and running
`make`. There is a arm64 target and an amd64 target. The default is
to build the arm64 target; to build the amd64 target run
`make rclone-report-amd64`.

There is some default configuration for the command line arguments
of the application that have generic defaults in the Makefile. To
customize these export the variables before running make and your local
configuration will be embedded in the resulting binary. The variables
are:

 * `VAULT_MATERIAL` - a path to a Vault Key/Value material that contains
   the B2 secret. It is expected that the material contains an `id` and
   `key` field with credentials for a Backblaze B2 account This assumes
   that the Key/Value store is mounted to the path kv/. This is optional
   but if it is not specified then password must be specified.
 * `B2_BUCKET` - the name of the target B2 bucket
 * `INSTANCE_NAME` - the instance name label attached to the prometheus
    metrics.
 * `PUSH_GATEWAY` - a full HTTP/S URL to the pushgateway instance

If these fields are not specified at build time they can be overriden as
command line flags.

## Running

The application supports a pre-environment file which is a JSON file
that contains key/value pairs that will be read and injected into the
process environment immediately after flag parsing but before any other
program logic. While the file can contain any environment variables
this is mostly designed for injecting Vault secrets. The default file
location is `/etc/default/restic-backup.json` but can be overridden
using the `--env-file` command line flag. If the specified file does not
exist it is silently ignored, any other errors reading or parsing the
file are fatal.

For example, this file configures the Vault JWT:

```json
{
    "VAULT_JWT": "a-jwt-token"
}
```

This program is designed to be run as a cron job. The code expects to be
able to fetch credentials from a Hashicorp Vault instance which must be
configured in the environment. One set of these variables is mandatory
and the application will fail to run without them:

This set configured AppRole authentication:

* `VAULT_ADDR` the HTTP/S address the Vault server.
* `VAULT_TOKEN` (optional) a Vault token to use for authentication
* `VAULT_ROLE_ID` and `VAULT_SECRET_ID` (optional) used to authenticate
  to Vault using the AppRole backend. Either these or `VAULT_TOKEN` must
  be specified otherwise Vault will fail to initialize.

This set configures JWT authentication:

* `VAULT_JWT` the JWT used to authenticate to Vault

On success the application will print noting and exit with a `0` status
code. Use prometheus to scrape the metrics from your pushgateway
instance and set alerts.

By default a DNS SRV query for `_vault._tcp.<search_domain>` will
be executed to locate the Vault host. Normally `<search_domain>`
is the `search` line from your `/etc/resolv.conf` file, however
this can be overridden at runtime by setting `SEARCH_DOMAIN` in the
environment. This autodiscovery behavior can be disabled entirely with
the `--no-discover-vault` command line flag.

The program expects the first argument to be the path to backup. In lieu
of this argument the `RCLONE_TO_B2_FROM` environment variable can be
specified. Failure to specify at least one of this is an error.

## Metrics

The following metrics are exported, all are gauges:

 * `rclone_job_last_success_unixtime` - Last time a batch job successfully finished
 * `rclone_error_count` - Number of errors encountered by rclone
 * `rclone_check_count` - Number of checks performed by rclone
 * `rclone_check_total_count` - Total number of checks to be performed by rclone
 * `rclone_transfers_count` - Number of transfers performed by rclone
 * `rclone_transfers_total_count` - Total number of transfers to be performed by rclone
 * `rclone_deleted_dirs` - Number of directories deleted by rclone
 * `rclone_deleted_files` - Number of files deleted by rclone
 * `rclone_renamed_files` - Number of files renamed by rclone
 * `rclone_elapsed_time` - Elapsed time that rclone has run, in milliseconds
 * `rclone_transfer_speed` - Transfer speed for rclone, in bytes/second
 * `rclone_transfer_bytes` - Number of bytes transferred by rclone
 * `rclone_transfer_total_bytes` - Total number of bytes to be transferred by rclone 

## Contributing

Contributions are welcomed. Please file a pull request and we'll
consider your changes. Please try to follow the style of the existing
code and do not add additional libraries without justification.

While we appreciate the time and effort of contributors there's not
guarantee that we'll be able to accept all contributions. If you're
interested in making a rather large change then please open an issue
first so we can discuss the implications of the change before you invest
too much time in making those changes.
