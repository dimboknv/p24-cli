# p24

`p24` is command line interface for communication with [privat24 marchant information api](https://api.privatbank.ua/#p24/main).

**Note:** before using `p24` you need to [register merchant in privat24 system](https://api.privatbank.ua/#p24/registration).

## Features

- getting merchant card balance

- getting merchant statements list for intervals greater than 90 days

- piping

- loading progress bar

- rate limiting and retrying

- export merchant statements list to a `xml|xlsx` encoding

- export only needed fields by `--format` options

## Installation

### go

```sh
go install github.com/dimboknv/p24-cli
p24-cli version
```

### build from source

```sh
git clone https://github.com/dimboknv/p24-cli.git && cd p24-cli
make build
./bin/p24 version
```

### docker image

```sh
git clone https://github.com/dimboknv/p24-cli.git && cd p24-cli
make docker 
docker run -it --rm p24 p24 version
```

## Usage

```bash
Usage:
  p24 [OPTIONS] statements [statements-OPTIONS]

Application Options:
      --debug                   Is debug mode?

Help Options:
  -h, --help                    Show this help message

[statements command options]
          --id=                 Merchant id
          --pass=               Merchant password
          --card=               Merchant card number
          --timeout=            http request timeout (default: 90s)
          --sd=                 Start date of statements date range with "dd.mm.yyyy" layout
          --ed=                 End date of statements date range with "dd.mm.yyyy" layout
      -f, --format=             Export format todo (default:
                                Card|Appcode|TranDate|Amount|CardAmount|Rest|Terminal|Description|,)
      -e, --encoding=[xml|xlsx] Export encoding (default: xml)
      -o, --out=                Export statements list to a file with specified extname encoding. If
                                empty export to stdout with '-e' encoding
```

## Piping

You can use `p24` in pipeline:

```sh
p24 statements --id="id" --pass="pass" --card="card" --sd="01.01.2022" --ed="01.02.2022" --timeout=10s --encoding=xml | dasel -p xml
```

## Docker

```sh
docker run --rm -it -v ${PWD}:/p24 p24 p24 statements --id="id" --pass="pass" --card="card" --sd="01.01.2022" --ed="01.02.2022" --timeout=10s --out=out.xlsx
```
