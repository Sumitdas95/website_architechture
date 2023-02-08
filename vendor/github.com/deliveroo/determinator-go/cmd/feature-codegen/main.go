package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

type feature struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type method struct {
	Name        string
	Description string
}

type featureData struct {
	Name                    string
	Tags                    []string
	IsEnabledMethod         method
	IsEnabledForActorMethod method
	VariantForActorMethod   method
	ConstantName            string
	ConstantDescription     string
}

type args struct {
	src    string
	dest   string
	pkg    string
	verify bool
	prefix string
}

func main() {
	log.SetFlags(0)
	var args args
	flag.StringVar(&args.src, "src", "./.florence/features.yml", "path to features.yml")
	flag.StringVar(&args.dest, "dest", "./internal/florence", "package destination")
	flag.StringVar(&args.pkg, "pkg", "florence", "package name")
	flag.StringVar(&args.prefix, "prefix", "", "prefix for feature flags to remove from the flag name. Multiple can be provided by comma-separating")
	flag.BoolVar(&args.verify, "verify", false, "exit with error and do nothing if out of sync")

	flag.Parse()

	data, err := ioutil.ReadFile(args.src)
	if err != nil {
		log.Fatalln("unable to open file:", err)
	}
	var rawFeatures map[string]feature
	if err := yaml.Unmarshal(data, &rawFeatures); err != nil {
		log.Fatalln("unable to unmarshal yml:", err)
	}

	features := make([]featureData, 0, len(rawFeatures))
	for name, f := range rawFeatures {
		flagDescription := f.Description
		tags := f.Tags
		if len(tags) > 0 {
			sort.Slice(tags, func(i, j int) bool {
				return tags[i] < tags[j]
			})
			flagDescription += "\ntags: [" + strings.Join(tags, ", ") + "]"
		}
		flagName := name
		for _, prefix := range strings.Split(args.prefix, ",") {
			flagName = strings.TrimPrefix(flagName, prefix)
		}
		flagName = camelCase(flagName)

		constantName := "FeatureFlag" + flagName

		features = append(
			features, featureData{
				Name: name,
				Tags: f.Tags,
				IsEnabledMethod: method{
					Name:        flagName + "Flag",
					Description: StringToGoComment(flagName + "Flag determinates if the feature flag is on, regardless of the Actor.\n " + flagDescription),
				},
				IsEnabledForActorMethod: method{
					Name:        flagName + "FlagForActor",
					Description: StringToGoComment(flagName + "FlagForActor determinates if the feature flag is on for a given Actor.\n" + flagDescription),
				},
				VariantForActorMethod: method{
					Name:        flagName + "VariantForActor",
					Description: StringToGoComment(flagName + "VariantForActor determinates the variant for a given Actor.\n" + flagDescription),
				},
				ConstantName:        constantName,
				ConstantDescription: StringToGoComment(constantName + " " + flagDescription),
			},
		)
	}
	sort.Slice(features, func(i, j int) bool {
		return features[i].Name < features[j].Name
	})

	tm := template.Must(template.New("flags.go").Parse(featuresTemplate))
	vars := struct {
		Features    []featureData
		PackageName string
		Src         string
	}{
		Features:    features,
		PackageName: args.pkg,
		Src:         args.src,
	}
	var b bytes.Buffer
	must(tm.Execute(&b, vars))

	err = os.MkdirAll(args.dest, os.ModePerm)
	must(err)

	destpath := filepath.Join(args.dest, "flags.go")
	writeGoFile(destpath, &b, args.verify)
}

func writeGoFile(path string, b *bytes.Buffer, verify bool) {
	formatted, err := format.Source(b.Bytes())
	must(err)
	if verify {
		destdata, err := ioutil.ReadFile(path)
		must(err)
		if !reflect.DeepEqual(formatted, destdata) {
			log.Fatalf("flags.go is a generated file by this script.\n")
			log.Fatalf("To add a flag, add it to features.yml and re-run this script.\n")
			log.Fatalf("%s is out of sync.\n", path)
			os.Exit(1)
		}
	}

	must(ioutil.WriteFile(path, formatted, 0644))
}

//go:embed features.tmpl
var featuresTemplate string

var camelCaseRegExp = regexp.MustCompile(`([^\w+]|[_])`)

var caser = cases.Title(language.BritishEnglish)

func camelCase(s string) string {
	parts := camelCaseRegExp.Split(s, -1)
	for i := range parts {
		parts[i] = caser.String(parts[i])
	}
	return strings.Join(parts, "")
}

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// StringToGoComment renders a possible multi-line string as a valid Go-Comment.
// Each line is prefixed as a comment.
// Apache 2.0 via oapi-codegen
func StringToGoComment(in string) string {
	if len(in) == 0 || len(strings.TrimSpace(in)) == 0 { // ignore empty comment
		return ""
	}

	// Normalize newlines from Windows/Mac to Linux
	in = strings.Replace(in, "\r\n", "\n", -1)
	in = strings.Replace(in, "\r", "\n", -1)

	// Add comment to each line
	var lines []string
	for _, line := range strings.Split(in, "\n") {
		lines = append(lines, fmt.Sprintf("// %s", line))
	}
	in = strings.Join(lines, "\n")

	// in case we have a multiline string which ends with \n, we would generate
	// empty-line-comments, like `// `. Therefore remove this line comment.
	in = strings.TrimSuffix(in, "\n// ")
	return in
}
