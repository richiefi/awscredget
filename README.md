awscredget
----------

`awscredget` is a minimal tool to acquire AWS session/role credentials. It is
useful for acquiring temporary AWS credentials to pass to short-lived containers
and similar isolated environments, particularly in cases where installing the
full `awscli` package on the host system is problematic.

## Building

The tool may be built for the current system with either `make` or `go build`.
To build for multiple architectures, `make all` may be used; set `ARCHS` in the
Makefile to add new targets.

## Usage

Run the tool with the `-h` flag to see all options. Typical usage from a shell
script:

    eval $(awscredget)
  
The `-r` flag may be used to acquire credentials for the given role ARN instead
of session credentials.

Input credentials are fetched using AWS SDK's default mechanisms.
