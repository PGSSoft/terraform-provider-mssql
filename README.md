# Terraform provider for MS SQL Server

- [Usage docs](docs/index.md)
- [Examples](examples/)

## Requirements
- [Terraform](https://www.terraform.io/downloads) >=1.0.0
- MS SQL Server >=2017 or Azure SQL 
- [Go](https://go.dev/doc/install) 1.19 (to build the provider)

## Building the provider 
Clone repository to: `$GOPATH/src/github.com/PGSSoft/terraform-provider-mssql`

```shell
$ cd $GOPATH/src/github.com/PGSSoft/terraform-provider-mssql
$ make install
```

## Acceptance tests 
Acceptance tests run MS SQL Server docker container automatically. 
Make sure you have Docker installed and configured before running the tests (i.e. make sure you're able to run `docker run <image>`).

To run all tests, including acceptance tests, using Docker MSSQL image:
```shell
$ make dockerAccTest
```

To run all tests, including acceptance tests, using Azure SQL:
```shell
$ az login
$ make azureAccTest
```


To run only unit tests (excluding tests depending on Docker container or Azure SQL):
```shell
$ make test
```

## About

The project maintained by [software development agency](https://www.pgs-soft.com/) [PGS Software](https://www.pgs-soft.com/).
See our other [open-source projects](https://github.com/PGSSoft) or [contact us](https://www.pgs-soft.com/contact-us/) to develop your product.


## Follow us

[![Twitter URL](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/intent/tweet?text=https://github.com/PGSSoft/terraform-provider-mssql)
[![Twitter Follow](https://img.shields.io/twitter/follow/pgssoftware.svg?style=social&label=Follow)](https://twitter.com/pgssoftware)