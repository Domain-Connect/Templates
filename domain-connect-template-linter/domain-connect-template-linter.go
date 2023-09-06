package main

/*
 * This is a Domain Connect template lint tool to validate contents of a template file.
 * These templates are usually found from https://github.com/domain-connect/templates
 *
 * Questions about the tool can be sent to <domain-connect@cloudflare.com>
 */

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Command-line options
var (
	Cloudflare  *bool
	PrettyPrint *bool
	Inplace     *bool
)

const (
	Max16b = 1<<16 - 1
	Max31b = 1<<31 - 1
)

type Template struct {
	ProviderID          string  `json:"providerId" validate:"required,min=1,max=64"`
	ProviderName        string  `json:"providerName" validate:"required,min=1,max=64"`
	ServiceID           string  `json:"serviceId" validate:"required,min=1,max=64"`
	ServiceName         string  `json:"serviceName" validate:"required,min=1,max=255"`
	Version             int     `json:"version,omitempty"`
	Logo                string  `json:"logoUrl,omitempty", validate:"url"`
	Description         string  `json:"description,omitempty"`
	VariableDescription string  `json:"variableDescription,omitempty"`
	Shared              bool    `json:"shared,omitempty"` /* deprecated */
	SyncBlock           bool    `json:"syncBlock,omitempty"`
	SharedProviderName  bool    `json:"sharedProviderName,omitempty"`
	SharedServiceName   bool    `json:"sharedServiceName,omitempty"`
	SyncPubKeyDomain    string  `json:"syncPubKeyDomain,omitempty" validate:"max=255"`
	SyncRedirectDomain  string  `json:"syncRedirectDomain,omitempty"`
	MultiInstance       bool    `json:"multiInstance,omitempty"`
	WarnPhishing        bool    `json:"warnPhishing,omitempty"`
	HostRequired        bool    `json:"hostRequired,omitempty"`
	Records             Records `json:"records"`
}

type Records []Record

type Record struct {
	Type      string `json:"type" validate:"required,min=1,max=16"`
	GroupID   string `json:"groupId,omitempty" validate:"min=1"`
	Essential string `json:"essential,omitempty" validate:"oneof=Always OnApply"`
	Host      string `json:"host" validate:"min=1,max=255"`
	Name      string `json:"name,omitempty" validate:"hostname"`
	PointsTo  string `json:"pointsTo,omitempty" validate:"min=1,max=255"`
	TTL       int    `json:"ttl" validate:"required,min=1"`
	Data      string `json:"data,omitempty" validate:"min=1"`
	TxtCMM    string `json:"txtConflictMatchingMode,omitempty" validate="oneof=None All Prefix"`
	TxtCMP    string `json:"txtConflictMatchingPrefix,omitempty" validate:"min=1"`
	Priority  int    `json:"priority,omitempty"`
	Weight    int    `json:"weight,omitempty"`
	Port      int    `json:"port,omitempty"`
	Protocol  string `json:"protocol,omitempty" validate:"min=1"`
	Service   string `json:"service,omitempty" validate:"min=1"`
	Target    string `json:"target,omitempty" validate:"min=1"`
	SPFRules  string `json:"spfRules,omitempty" validate:"min=1"`
}

func setLoglevel(loglevel string) {
	level, err := zerolog.ParseLevel(loglevel)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid loglevel")
	}
	zerolog.SetGlobalLevel(level)
}

func main() {
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		},
	)
	Cloudflare = flag.Bool("cloudflare", false, "use Cloudflare specific template rules")
	PrettyPrint = flag.Bool("pretty", false, "pretty-print template json")
	Inplace = flag.Bool("inplace", false, "inplace write back pretty-print")
	loglevel := flag.String("loglevel", "info", "loglevel can be one of: panic fatal error warn info debug trace")
	flag.Parse()

	setLoglevel(*loglevel)

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <template.json> [template2.json ...]\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatal().Msg("template file(s) as command argument is missing")
	}

	exitVal := 0
	for _, arg := range flag.Args() {
		exitVal |= checkTemplate(arg)
	}
	os.Exit(exitVal)
}

func checkTemplate(templatePath string) int {
	retVal := 0
	tlog := log.With().Str("template", templatePath).Logger()

	f, err := os.Open(templatePath)
	if err != nil {
		tlog.Error().Err(err).Msg("cannot open file")
		return 1
	}
	defer f.Close()

	var r io.Reader
	r = f

	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var template Template
	err = decoder.Decode(&template)

	if err != nil {
		tlog.Error().Err(err).Msg("json decode error")
		retVal = 1
	}

	Validator := validator.New()
	err = Validator.Struct(template)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			tlog.Warn().Err(err).Msg("template field validation")
			retVal = 1
		}
	}

	if template.Version < 0 {
		tlog.Info().Msg("use of negative version number is not recommended")
	}

	if template.Shared {
		tlog.Info().Msg("shared flag has been deprecated, use sharedProviderName instead")
		// switch to false so that deprecated flag is not part of
		// pretty printed output
		template.Shared = false
	}

	if *Cloudflare {
		retVal |= CloudflareTemplateChecks(tlog, template)
	}

	for rnum, record := range template.Records {
		retVal |= checkRecord(tlog, rnum, record)
	}

	if *PrettyPrint || *Inplace {
		pp, err := json.Marshal(template)
		if err != nil {
			tlog.Error().Err(err).Msg("")
		}
		var out bytes.Buffer
		json.Indent(&out, pp, "", "    ")
		fmt.Fprintf(&out, "\n")
		if *Inplace {
			WriteBack(tlog, templatePath, out)
		} else {
			out.WriteTo(os.Stdout)
		}
	}

	return retVal
}

func WriteBack(tlog zerolog.Logger, templatePath string, out bytes.Buffer) {
	f, err := os.CreateTemp("./", path.Base(templatePath))
	if err != nil {
		tlog.Warn().Err(err).Msg("could not create temporary file")
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	out.WriteTo(w)
	if err != nil {
		tlog.Warn().Err(err).Msg("could write to temporary file")
		return
	}
	w.Flush()
	err = os.Rename(f.Name(), templatePath)
	if err != nil {
		tlog.Warn().Err(err).Msg("could not move template back inplace")
	}
	tlog.Debug().Str("tmpfile", f.Name()).Msg("updated")
}

func checkRecord(tlog zerolog.Logger, rnum int, record Record) int {
	rlog := tlog.With().Int("record", rnum).Logger()
	retVal := 0
	switch record.Type {
	case "CNAME", "NS":
		if record.Host == "@" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be @")
			retVal = 1
		}
		fallthrough
	case "A", "AAAA":
		if record.Host == "" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be empty")
			retVal = 1
		}
		if record.TTL < 0 || Max31b < record.TTL {
			rlog.Error().Str("type", record.Type).Int("ttl", record.TTL).Msg("invalid TTL")
			retVal = 1
		}

	case "TXT":
		if record.Host == "" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be empty")
			retVal = 1
		}
		if record.Data == "" {
			rlog.Error().Str("type", record.Type).Msg("record data must not be empty")
			retVal = 1
		}
		if record.TTL < 0 || Max31b < record.TTL {
			rlog.Error().Str("type", record.Type).Int("ttl", record.TTL).Msg("invalid TTL")
			retVal = 1
		}
		if *Cloudflare {
			if record.TxtCMM != "" || record.TxtCMM == "None" {
				rlog.Info().Msg("Cloudflare does not support txtConflictMatchingMode record settings")
			}
			if record.TxtCMP != "" {
				rlog.Info().Msg("Cloudflare does not support txtConflictMatchingPrefix record settings")
			}
		} else {
			if record.TxtCMM == "Prefix" && record.TxtCMP == "" {
				rlog.Warn().Str("type", record.Type).Msg("record txtConflictMatchingPrefix is not defined")
				retVal = 1
			}
		}

	case "MX":
		if record.Host == "" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be empty")
			retVal = 1
		}
		if record.PointsTo == "" {
			rlog.Error().Str("type", record.Type).Msg("record pointsTo must not be empty")
			retVal = 1
		}
		if record.TTL < 0 || Max31b < record.TTL {
			rlog.Error().Str("type", record.Type).Int("ttl", record.TTL).Msg("invalid TTL")
			retVal = 1
		}
		if record.Priority < 0 || Max31b < record.Priority {
			rlog.Error().Str("type", record.Type).Int("priority", record.Priority).Msg("invalid priority")
			retVal = 1
		}

	case "SRV":
		if record.Name == "" {
			rlog.Error().Str("type", record.Type).Msg("record name must not be empty")
			retVal = 1
		}
		if record.Target == "" {
			rlog.Error().Str("type", record.Type).Msg("record target must not be empty")
			retVal = 1
		}
		if record.Priority < 0 || Max31b < record.Priority {
			rlog.Error().Str("type", record.Type).Int("priority", record.Priority).Msg("invalid priority")
			retVal = 1
		}
		if record.Service == "" {
			rlog.Error().Str("type", record.Type).Msg("record service must not be empty")
			retVal = 1
		}
		if record.Weight < 0 || Max31b < record.Weight {
			rlog.Error().Str("type", record.Type).Int("weigth", record.Weight).Msg("invalid weigth")
			retVal = 1
		}
		if record.Port < 1 || Max16b < record.Port {
			rlog.Error().Str("type", record.Type).Int("port", record.Port).Msg("invalid port")
			retVal = 1
		}

	case "SPFM":
		if record.Host == "" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be empty")
			retVal = 1
		}
		if record.SPFRules == "" {
			rlog.Error().Str("type", record.Type).Msg("record spfRules must not be empty")
			retVal = 1
		}
	}

	if strings.Count(record.GroupID, "%") > 1 {
		rlog.Error().Msg("record groupId must not be variable")
		retVal = 1
	}

	if strings.Count(record.TxtCMP, "%") > 1 {
		rlog.Error().Msg("record txtConflictMatchingPrefix must not be variable")
		retVal = 1
	}

	if *Cloudflare {
		if record.Essential != "" {
			rlog.Info().Msg("Cloudflare does not support essential record settings")
		}
	}

	return retVal
}

func CloudflareTemplateChecks(tlog zerolog.Logger, template Template) int {
	retVal := 0
	if template.SyncBlock {
		tlog.Error().Msg("Cloudflare does not support syncBlock")
		retVal = 1
	}
	if template.SyncPubKeyDomain == "" {
		tlog.Error().Msg("Cloudflare requires syncPubKeyDomain")
		retVal = 1
	}
	if template.SharedServiceName {
		tlog.Info().Msg("Cloudflare does not support sharedServiceName")
	}
	if template.SyncRedirectDomain != "" {
		tlog.Info().Msg("Cloudflare does not support syncRedirectDomain")
	}
	if template.MultiInstance {
		tlog.Info().Msg("Cloudflare does not support multiInstance")
	}
	if template.WarnPhishing {
		tlog.Info().Msg("Cloudflare does not use warnPhishing because syncPubKeyDomain is required")
	}
	if template.HostRequired {
		tlog.Info().Msg("Cloudflare does not support hostRequired")
	}
	return retVal
}
