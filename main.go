package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed katas.yaml
var defaultKatas []byte

func usage() {
	fmt.Fprintf(os.Stderr, "usage: katas [options]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	doneFlag = flag.String("done", "", "mark `kata` as done today")
	initFlag = flag.Bool("init", false, "initialize config at "+configPath())
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("katas: ")

	flag.Usage = usage
	flag.Parse()

	katas := newKatas()

	if *initFlag {
		if err := katas.initConfig(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *doneFlag != "" {
		if err := katas.markDone(*doneFlag); err != nil {
			log.Fatal(err)
		}
		return
	}

	katas.print()
}

// kata represents a programming exercise.
type kata struct {
	Name string   `yaml:"name"`
	URL  string   `yaml:"url"`
	Done []string `yaml:"done,omitempty"`
}

// katas represents the status of your programming training.
type katas struct {
	configPath string
	katas      []kata
}

func configPath() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".katas.yaml")
}

func newKatas() *katas {
	return &katas{configPath: configPath()}
}

func (k *katas) initConfig() error {
	if _, err := os.Stat(k.configPath); err == nil {
		return fmt.Errorf("config file %s already exists", k.configPath)
	}
	return os.WriteFile(k.configPath, defaultKatas, 0644)
}

// load reads the katas from the config file
func (k *katas) load() error {
	data, err := os.ReadFile(k.configPath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &k.katas); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	return nil
}

// save writes the katas to the config file
func (k *katas) save() error {
	data, err := yaml.Marshal(k.katas)
	if err != nil {
		return fmt.Errorf("marshaling katas: %w", err)
	}

	if err := os.WriteFile(k.configPath, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// print displays all katas with their status
func (k *katas) print() error {
	if err := k.load(); err != nil {
		return err
	}

	if len(k.katas) == 0 {
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	defer tw.Flush()

	format := "%v\t%v\t%v\t%v\n"

	fmt.Fprintf(tw, format, "Name", "Done", "Last done", "URL")
	fmt.Fprintf(tw, format, "----", "----", "---------", "---")

	var totalTimesDone TimesDone
	var latestLastDone LastDone

	for _, kata := range k.katas {
		timesDone := TimesDone(len(kata.Done))
		totalTimesDone += timesDone

		var lastDone LastDone
		for _, d := range kata.Done {
			t, err := time.Parse("2006-01-02", d)
			if err != nil {
				log.Printf("parsing kata %q in %s: %v", kata.Name, configPath(), err)
				continue
			}
			if t.After(lastDone.t) {
				lastDone.t = t
			}
			if t.After(latestLastDone.t) {
				latestLastDone.t = t
			}
		}

		fmt.Fprintf(tw, format, kata.Name, timesDone, lastDone, kata.URL)
	}

	fmt.Fprintf(tw, format, "----", "----", "---------", "---")
	fmt.Fprintf(tw, format, len(k.katas), totalTimesDone, latestLastDone, "")

	return nil
}

type LastDone struct {
	t time.Time
}

func (ld LastDone) String() string {
	if ld.t.IsZero() {
		return "never"
	}
	daysAgo := int(time.Since(ld.t).Hours() / 24)
	return fmt.Sprintf("%d days ago", daysAgo)
}

type TimesDone int

func (d TimesDone) String() string {
	return fmt.Sprintf("%dx", d)
}

// markDone marks a kata as completed today
func (k *katas) markDone(name string) error {
	if err := k.load(); err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")

	for i, kata := range k.katas {
		if kata.Name == name {
			// Check if already done today
			if len(kata.Done) > 0 && kata.Done[len(kata.Done)-1] == today {
				return fmt.Errorf("kata %s already marked as done today\n", name)
			}

			k.katas[i].Done = append(kata.Done, today)
			return k.save()
		}
	}

	return fmt.Errorf("kata %s not found in %s", name, k.configPath)
}
