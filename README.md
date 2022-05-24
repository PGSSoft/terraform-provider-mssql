# Terraform provider for MS SQL Server

- [Usage docs](docs/index.md)
- [Examples](examples/)

## Requirements
- [Terraform](https://www.terraform.io/downloads) >=1.0.0
- MS SQL Server >=2017 or Azure SQL 
- [Go](https://go.dev/doc/install) 1.18 (to build the provider)

## Building the provider 
Clone repository to: `$GOPATH/src/github.com/tkielar/terraform-provider-mssql`

```shell
$ cd $GOPATH/src/github.com/tkielar/terraform-provider-mssql
$ go install 
```

## Acceptance tests 
Acceptance tests run MS SQL Server docker container automatically. 
Make sure you have Docker installed and configured before running the tests (i.e. make sure you're able to run `docker run <image>`).

To run all tests, including acceptance tests:
```shell
$ TF_ACC=1 go test -v ./...
```

To run only unit tests (excluding tests depending on Docker container):
```shell
$ go test -v ./...
```