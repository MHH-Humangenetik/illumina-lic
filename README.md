# illumina-lic

CLI to query Illumina DRAGEN license quota usage and print a table with limit, usage, and expiration per pipeline.

## Setup

```sh
cp .env.example .env   # fill in LICENSE_USERNAME and LICENSE_PASSWORD
go build -o illumina-lic .
```

## Usage

```sh
./illumina-lic                  # table with pipeline quota usage
./illumina-lic -info            # also show message and credential expiration
./illumina-lic -endpoint URL    # override license API endpoint
```

## Test

```sh
go test ./...
```
