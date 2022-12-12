# urlbits

Pipe in a bunch of urls and see their bits.

[![Go Report Card](https://goreportcard.com/badge/github.com/davemolk/urlbits)](https://goreportcard.com/report/github.com/davemolk/urlbits)
## Overview
The default behavior is to parse the urls and print their components to Stdout as json. Use a flag to narrow the output to a specific component.

## Examples
```
$ cat urls.txt
https://www.google.com/search?name=golang&language=en&mascot=gopher
https://go.dev/play/
google.com
foo/bar
postgres://user:pass@host.com:5432/path1?k=v#f
totaljunk
```

### Default Usage
```
$ cat urls.txt | urlbits
{
  "scheme": "https",
  "host": "www.google.com",
  "path": "/search",
  "raw_query": "name=golang\u0026language=en\u0026mascot=gopher"
}
{
  "scheme": "https",
  "host": "go.dev",
  "path": "/play/"
}
{
  "scheme": "postgres",
  "user": "user:pass",
  "host": "host.com:5432",
  "path": "/path1",
  "raw_query": "k=v#f"
}
```

### Keys
```
$ cat urls.txt | urlbits -keys
name
language
mascot
k
```

### Paths
```
$ cat urls.txt | urlbits -paths
/search
/play/
/path1
```

## Install
First, you'll need to [install go](https://golang.org/doc/install). Then, run the following command:

```
go install github.com/davemolk/urlbits
```

## Command-line Options
```
Usage of urlbits:
  -domains bool
    	Output the domains.
  -keys bool
    	Output the keys.
  -kv bool
    	Output keys and values.
  -paths bool
    	Output the paths.
  -save bool
    	Save output to a file.
  -user bool
    	Output user information (username and password).
  -values bool
    	Output the values.
  -validate bool
    	Strip out URLs without a scheme and host.
  -verbose bool
    	Verbose output.
```

### Why pipelines?
While they do work well in this scenario, I was trying to think of ways to practice using pipelines in Go that were a little more practical than what I was finding in tutorials and books. 