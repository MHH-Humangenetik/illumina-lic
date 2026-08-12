# illumina-lic

CLI to query Illumina DRAGEN license quota usage and print a table with limit, usage, and expiration per pipeline.

## Example Output

```sh
┌─────────────────────────────────────────────────────────────────────────────────┐
| Pipeline    | Limit        | Used          | % Used | Expires                   |
| ─────────────────────────────────────────────────────────────────────────────── |
| Compression | 10.00 Tbases | 0.0000 Tbases | 0.00%  | 2027-07-28T00:00:00+00:00 |
| Genome      | 10.00 Tbases | 0.2786 Tbases | 2.79%  | 2027-07-28T00:00:00+00:00 |
└─────────────────────────────────────────────────────────────────────────────────┘
```

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
