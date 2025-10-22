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
	initFlag = flag.Bool("init", false, "initialize "+katasFilePath())
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
	}
	katas.print(*doneFlag)
}

// kata represents a programming exercise.
type kata struct {
	Name string   `yaml:"name"`
	URL  string   `yaml:"url"`
	Done []string `yaml:"done,omitempty"`
}

// katas represent a programming training.
type katas struct {
	filePath string
	katas    []kata
}

func katasFilePath() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".katas.yaml")
}

func newKatas() *katas {
	return &katas{filePath: katasFilePath()}
}

func (k *katas) initConfig() error {
	if _, err := os.Stat(k.filePath); err == nil {
		return fmt.Errorf("file %s already exists", k.filePath)
	}
	return os.WriteFile(k.filePath, defaultKatas, 0644)
}

// load reads katas from the file.
func (k *katas) load() error {
	data, err := os.ReadFile(k.filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &k.katas); err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}

	return nil
}

// save writes katas to the file.
func (k *katas) save() error {
	data, err := yaml.Marshal(k.katas)
	if err != nil {
		return fmt.Errorf("marshaling katas: %w", err)
	}

	if err := os.WriteFile(k.filePath, data, 0644); err != nil {
		return err
	}

	return nil
}

// print displays all katas with their status.
func (k *katas) print(doneKata string) error {
	if err := k.load(); err != nil {
		return err
	}

	if len(k.katas) == 0 {
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	defer tw.Flush()

	format := "%v\t%v\t%v\t%v\t%v\n"

	fmt.Fprintf(tw, format, "Name", "Done", "Last done", "Mastery", "URL")
	fmt.Fprintf(tw, format, "----", "----", "---------", "-------", "---")

	var totalTimesDone TimesDone
	var latestLastDone LastDone
	var totalMastery Mastery

	for _, kata := range k.katas {
		timesDone := TimesDone(len(kata.Done))
		totalTimesDone += timesDone

		var lastDone LastDone
		for _, d := range kata.Done {
			t, err := time.Parse("2006-01-02", d)
			if err != nil {
				log.Printf("parsing kata %q in %s: %v", kata.Name, katasFilePath(), err)
				continue
			}
			if t.After(lastDone.t) {
				lastDone.t = t
			}
			if t.After(latestLastDone.t) {
				latestLastDone.t = t
			}
		}

		kataName := kata.Name
		if kataName == doneKata {
			kataName = "> " + kataName
		}

		kataMastery := mastery(int(timesDone), lastDone.t)
		totalMastery += kataMastery

		fmt.Fprintf(tw, format, kataName, timesDone, lastDone, kataMastery, kata.URL)
	}

	fmt.Fprintf(tw, format, "----", "----", "---------", "-------", "---")
	var avgMastery Mastery
	if len(k.katas) > 0 {
		avgMastery = Mastery(int(totalMastery) / len(k.katas))
	} else {
		avgMastery = 0
	}
	fmt.Fprintf(tw, format, len(k.katas), totalTimesDone, latestLastDone, avgMastery, "")

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

func (td TimesDone) String() string {
	return fmt.Sprintf("%dx", td)
}

// markDone marks a kata as completed today.
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

	return fmt.Errorf("kata %s not found in %s", name, k.filePath)
}
