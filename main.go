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

//go:embed default_katas.yaml
var defaultKatas []byte

func usage() {
	fmt.Fprintf(os.Stderr, "usage: katas [options]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	doneFlag = flag.String("done", "", "mark `kata` as done today")
	initFlag = flag.Bool("init", false, "initialize config with default katas at "+configPath())
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
	return filepath.Join(h, ".config", "katas.yaml")
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "Kata\tLast done\tDone\tURL\n")
	fmt.Fprintf(w, "----\t---------\t----\t---\n")

	totalDone := 0
	for _, kata := range k.katas {
		lastDone := "never"
		doneCount := len(kata.Done)
		totalDone += doneCount

		if doneCount > 0 {
			// Parse the last done date
			if lastDate, err := time.Parse("2006-01-02", kata.Done[doneCount-1]); err == nil {
				days := int(time.Since(lastDate).Hours() / 24)
				lastDone = fmt.Sprintf("%d days ago", days)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%dx\t%s\n", kata.Name, lastDone, doneCount, kata.URL)
	}

	fmt.Fprintf(w, "----\t\t----\t\n")
	fmt.Fprintf(w, "%d\t\t%dx\t\n", len(k.katas), totalDone)

	return nil
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
