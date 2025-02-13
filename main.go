package main

import (
	`flag`
	`fmt`
	`os`
	`text/tabwriter`
	`time`

	`github.com/beevik/ntp`
)

type ntpServers []string

func (s *ntpServers) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *ntpServers) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type ntpTime struct {
	Err error
	T   time.Time
	S   string
}

func main() {
	var ntpServer ntpServers
	flag.Var(&ntpServer, "s", "-s server1 -s server2:port2 ...")
	flag.Parse()

	if len(ntpServer) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	respChan := make(chan *ntpTime, len(ntpServer))
	for _, server := range ntpServer {
		go func(server string) {
			t, err := ntp.Time(server)
			if err != nil {
				respChan <- &ntpTime{
					Err: err,
					S:   server,
				}
			} else {
				respChan <- &ntpTime{
					T: t,
					S: server,
				}
			}
		}(server)
	}

	writer := tabwriter.NewWriter(os.Stdout, 2, 8, 1, '\t', tabwriter.AlignRight)
	failed := false
	for i := 0; i < len(ntpServer); i++ {
		resp := <-respChan
		if resp.Err != nil {
			fmt.Fprintf(writer, "[%s]\t\t\033[31mError\033[0m: %s\n", resp.S, resp.Err.Error())
			failed = true
		} else {
			fmt.Fprintf(writer, "[%s]\t\t\033[32mTime \033[0m: %s\n", resp.S, resp.T.Format(time.RFC3339))
		}
	}

	writer.Flush()

	if failed {
		os.Exit(1)
	}
}
