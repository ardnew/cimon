package config

import (
	"flag"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

//nolint:gochecknoglobals,gomnd
var (
	DefaultAddr          = Addr{port: 8080}
	DefaultProto         = TCP
	DefaultTimeout       = 5 * time.Second
	DefaultHeaderTimeout = 1 * time.Second
)

var ErrProtocol = errors.New("unknown protocol")

const anyInterfaceIPv4 = "0.0.0.0"

func New(out io.Writer, name, version string) *Flags {
	set := &Flags{
		FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		Bind:    DefaultAddr,
		Proto:   DefaultProto,
	}
	var rs string
	set.Usage, rs = set.usage(out, name, version)
	set.Var(&set.Bind, "l", "Bind to interface `[ADDR]:PORT`."+rs+
		"(Omit ADDR to listen on all interfaces.)")
	set.Var(&set.Proto, "p", "Use protocol `PROTO` (tcp, http).")
	set.Func("v", "Enable additional, verbose logging.",
		func(string) error { set.Verbose++; return nil },
	)
	set.DurationVar(&set.Timeout,
		"to", DefaultTimeout, "Timeout for HTTP content.")
	set.DurationVar(&set.HeaderTimeout,
		"th", DefaultHeaderTimeout, "Timeout for HTTP header.")
	return set
}

type Flags struct {
	*flag.FlagSet
	Bind          Addr
	Proto         Proto
	Timeout       time.Duration
	HeaderTimeout time.Duration
	Verbose       uint
}

func (f *Flags) usage(out io.Writer, name, version string) (func(), string) {
	const sep, tab = "\x30", 10
	type arg struct{ expr, desc string }
	args := map[string]arg{}
	args["h"] = arg{"-h", "Hi."}
	stop := max(tab, len(args["h"].expr))
	return func() {
		fmt.Fprintf(out, "%s version %s usage:\n", name, version)
		fmt.Fprintln(out)
		f.VisitAll(func(flg *flag.Flag) {
			name, usage := flag.UnquoteUsage(flg)
			expr := fmt.Sprintf("-%s %s", flg.Name, name)
			if len(expr) > stop {
				stop = len(expr)
			}
			args[flg.Name] = arg{expr, usage}
		})
		stop += 2
		keys := maps.Keys(args)
		sort.Strings(keys)
		for _, key := range keys {
			desc := strings.ReplaceAll(args[key].desc, sep,
				"\n   "+strings.Repeat(" ", stop+1))
			fmt.Fprintf(out, "  %-*s %s\n", stop, args[key].expr, desc)
		}
		fmt.Fprintln(out)
	}, sep
}

type Addr struct {
	string
	port uint
}

func (a *Addr) String() string {
	return a.string + ":" + strconv.FormatUint(uint64(a.port), 10)
}

func (a *Addr) Set(set string) error {
	n := strings.LastIndex(set, ":")
	if n >= 0 {
		sub := set[n+1:]
		set = strings.TrimSpace(set[:n])
		p, err := strconv.ParseUint(sub, 10, 16)
		if err != nil {
			return err
		}
		a.port = uint(p)
	}
	if set == "" {
		set = anyInterfaceIPv4
	}
	a.string = set
	return nil
}

type Proto uint

const (
	TCP Proto = iota
	HTTP
)

func (p *Proto) String() string {
	switch *p {
	case TCP:
		return "tcp"
	case HTTP:
		return "http"
	}
	return "unknown"
}

func (p *Proto) Set(set string) error {
	switch set {
	case "tcp":
		*p = TCP
	case "http":
		*p = HTTP
	default:
		return errors.Wrap(ErrProtocol, set)
	}
	return nil
}
