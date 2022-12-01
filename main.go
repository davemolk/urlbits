package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
)

type config struct {
	domains bool
	file    string
	keys    bool
	kv      bool
	paths   bool
	save    bool
	user    bool
	values  bool
	verbose bool
}

type urlbits struct {
	config config
}

// User field has been changed from *Userinfo found in the source code. 
type URL struct {
	Scheme      string `json:"scheme,omitempty"`
	Opaque      string `json:"opaque,omitempty"`    // encoded opaque data
	User        string  `json:"user,omitempty"` // username and password information
	Host        string `json:"host,omitempty"`   // host or host:port
	Path        string `json:"path,omitempty"`    // path (relative paths may omit leading slash)
	RawPath     string  `json:"raw_path,omitempty"`  // encoded path hint (see EscapedPath method)
	RawQuery    string `json:"raw_query,omitempty"`   // encoded query values, without '?'
	Fragment    string `json:"fragment,omitempty"`   // fragment for references, without '#'
	RawFragment string `json:"raw_fragment,omitempty"`   // encoded fragment hint (see EscapedFragment method)
}

func main() {
	var config config
	flag.BoolVar(&config.domains, "domains", false, "output domains")
	flag.StringVar(&config.file, "file", "", "name of file containing urls to parse")
	flag.BoolVar(&config.keys, "keys", false, "output keys")
	flag.BoolVar(&config.kv, "kv", false, "output keys and values")
	flag.BoolVar(&config.paths, "paths", false, "output paths")
	flag.BoolVar(&config.save, "save", true, "save output to file")
	flag.BoolVar(&config.user, "user", false, "output username and password")
	flag.BoolVar(&config.values, "values", false, "output values")
	flag.BoolVar(&config.verbose, "verbose", false, "verbose output")
	flag.Parse()

	ub := &urlbits{
		config: config,
	}

	ch, err := ub.read()
	if err != nil {
		log.Fatal("read failed", err)
	}

	var f *os.File
	if config.save {
		f, err = os.Create("results.txt")
		if err != nil {
			log.Fatal("unable to create file", err)
		}
		defer f.Close()
	}
	
	switch {
	case config.domains:
		for d := range ub.domains(ub.parsed(ch)) {
			fmt.Println(d)
			if config.save {
				ub.write(f, d)
			}
		}
	case config.keys:
		for k := range ub.keys(ub.kvMap(ub.parsed(ch))) {
			fmt.Println(k)
			if config.save {
				ub.write(f, k)
			}
		}
	case config.kv:
		for kv := range ub.kvMap(ub.parsed(ch)) {
			fmt.Println(kv)
			if config.save {
				data, err := json.Marshal(kv)
				if err != nil {
					log.Printf("unable to marshal data: %v", err)
					continue
				}
				ub.write(f, string(data))
			}
		}
	case config.paths:
		for p := range ub.paths(ub.parsed(ch)) {
			fmt.Println(p)
			if config.save {
				ub.write(f, p)
			}
		}
	case config.user:
		for u := range ub.user(ub.parsed(ch)) {
			fmt.Println(u)
			if config.save {
				ub.write(f, u.String())
			}
		}
	case config.values:
		for v := range ub.values(ub.kvMap(ub.parsed(ch))) {
			fmt.Println(v)
			if config.save {
				ub.write(f, v)
			}
		}
	default:
		for u := range ub.parsed(ch) {
			url := &URL{
				Scheme: u.Scheme,
				Opaque: u.Opaque,
				User: u.User.String(),
				Host: u.Host,
				Path: u.Path,
				RawPath: u.RawPath,
				RawQuery: u.RawQuery,
				Fragment: u.Fragment,
				RawFragment: u.RawFragment,
			}

			data, err := json.Marshal(url)
			if err != nil {
				log.Printf("unable to marshal: %v\n", err)
				continue
			}
			fmt.Println(string(data))
			ub.write(f, string(data))
		}
			
		
	}
}

func (ub *urlbits) read() (<-chan string, error) {
	ch := make(chan string)
	s := bufio.NewScanner(os.Stdin)

	go func(ch chan string) {
		defer close(ch)
		for s.Scan() {
			ch <- s.Text()
		}
		if err := s.Err(); err != nil && ub.config.verbose {
			log.Println(err)
		}
	}(ch)

	return ch, nil
}

func (ub *urlbits) parsed(urls <-chan string) <-chan *url.URL {
	ch := make(chan *url.URL)

	go func() {
		defer close(ch)
		for u := range urls {
			s, err := url.ParseRequestURI(u) // switch to Parse?
			if err != nil {
				if ub.config.verbose {
					log.Printf("parsing error for %s: %v\n", u, err)
				}
				continue
			}
			ch <- s
		}
	}()

	return ch
}

func (ub *urlbits) user(urls <-chan *url.URL) <-chan *url.Userinfo {
	ch := make(chan *url.Userinfo)

	go func() {
		defer close(ch)
		for u := range urls {
			if u.User != nil {
				ch <- u.User
			}
		}
	}()

	return ch
}

func (ub *urlbits) domains(urls <-chan *url.URL) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		for u := range urls {
			ch <- u.Host
		}
	}()

	return ch
}

func (ub *urlbits) paths(urls <-chan *url.URL) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		for u := range urls {
			if u.Path != "/" && u.Path != "" {
				ch <- u.Path
			}
		}
	}()

	return ch
}

func (ub *urlbits) kvMap(urls <-chan *url.URL) <-chan url.Values {
	ch := make(chan url.Values)

	go func() {
		defer close(ch)
		for u := range urls {
			m, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				if ub.config.verbose {
					log.Printf("param parsing error: %v\n", err)
				}
				continue
			}
			if len(m) > 0 {
				ch <- m
			}
		}
	}()

	return ch
}

func (ub *urlbits) keys(kvMap <-chan url.Values) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		for kv := range kvMap {
			for key := range kv {
				ch <- key
			}
		}
	}()

	return ch
}

func (ub *urlbits) values(kvMap <-chan url.Values) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		for kv := range kvMap {
			for _, value := range kv {
				for _, v := range value {
					ch <- v
				}
			}
		}
	}()

	return ch
}

func (ub *urlbits) write(f *os.File, s string) {
	_, err := fmt.Fprintf(f, "%v\n", s)
	if err != nil {
		log.Printf("error writing %s: %v\n", s, err)
	}
}