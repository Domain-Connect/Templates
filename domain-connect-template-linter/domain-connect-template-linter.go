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
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Command-line options
var (
	CheckLogos  *bool
	Cloudflare  *bool
	Inplace     *bool
	PrettyPrint *bool
)

const (
	Max16b = 1<<16 - 1
	Max31b = 1<<31 - 1
)

var exitVal int
var tlog zerolog.Logger
var collision map[string]bool

type Template struct {
	ProviderID          string  `json:"providerId" validate:"required,min=1,max=64"`
	ProviderName        string  `json:"providerName" validate:"required,min=1,max=64"`
	ServiceID           string  `json:"serviceId" validate:"required,min=1,max=64"`
	ServiceName         string  `json:"serviceName" validate:"required,min=1,max=255"`
	Version             SINT    `json:"version,omitempty"`
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
	Host      string `json:"host,omitempty" validate:"min=1,max=255"`
	Name      string `json:"name,omitempty" validate:"hostname"`
	PointsTo  string `json:"pointsTo,omitempty" validate:"min=1,max=255"`
	TTL       SINT   `json:"ttl" validate:"required,min=1"`
	Data      string `json:"data,omitempty" validate:"min=1"`
	TxtCMM    string `json:"txtConflictMatchingMode,omitempty" validate="oneof=None All Prefix"`
	TxtCMP    string `json:"txtConflictMatchingPrefix,omitempty" validate:"min=1"`
	Priority  SINT   `json:"priority,omitempty"`
	Weight    SINT   `json:"weight,omitempty"`
	Port      SINT   `json:"port,omitempty"`
	Protocol  string `json:"protocol,omitempty" validate:"min=1"`
	Service   string `json:"service,omitempty" validate:"min=1"`
	Target    string `json:"target,omitempty" validate:"min=1"`
	SPFRules  string `json:"spfRules,omitempty" validate:"min=1"`
}

type SINT int

func (sint *SINT) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		return json.Unmarshal(b, (*int)(sint))
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	tlog.Warn().Str("value", s).Msg("do not quote an integer, it makes it string")
	exitVal = 1
	*sint = SINT(i)
	return nil
}

func setLoglevel(loglevel string) {
	level, err := zerolog.ParseLevel(loglevel)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid loglevel")
	}
	zerolog.SetGlobalLevel(level)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <template.json> [...]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Warning. -inplace and -pretty will remove zero priority MX and SRV fields\n")
	}
	if isatty.IsTerminal(os.Stderr.Fd()) {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			},
		)
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	CheckLogos = flag.Bool("logos", false, "check logo urls are reachable (requires network)")
	Cloudflare = flag.Bool("cloudflare", false, "use Cloudflare specific template rules")
	Inplace = flag.Bool("inplace", false, "inplace write back pretty-print")
	PrettyPrint = flag.Bool("pretty", false, "pretty-print template json")
	loglevel := flag.String("loglevel", "info", "loglevel can be one of: panic fatal error warn info debug trace")
	flag.Parse()

	setLoglevel(*loglevel)

	if flag.NArg() < 1 {
		flag.Usage()
		log.Fatal().Msg("template file(s) command argument is missing")
	}

	collision = make(map[string]bool)

	for _, arg := range flag.Args() {
		checkTemplate(arg)
	}
	os.Exit(exitVal)
}

func checkTemplate(templatePath string) {
	tlog = log.With().Str("template", templatePath).Logger()

	f, err := os.Open(templatePath)
	if err != nil {
		tlog.Error().Err(err).Msg("cannot open file")
		exitVal = 1
		return
	}
	defer f.Close()
	tlog.Debug().Msg("processing template")

	var r io.Reader
	r = f

	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var template Template
	err = decoder.Decode(&template)

	if err != nil {
		tlog.Error().Err(err).Msg("json decode error")
		exitVal = 1
	}

	if checkInvalidChars(template.ProviderID) {
		tlog.Error().Str("providerId", template.ProviderID).Msg("providerId contains invalid characters")
		exitVal = 1
	}

	if checkInvalidChars(template.ServiceID) {
		tlog.Error().Str("serviceId", template.ServiceID).Msg("serviceId contains invalid characters")
		exitVal = 1
	}

	if _, found := collision[template.ProviderID+"/"+template.ServiceID]; found {
		tlog.Error().Str("providerId", template.ProviderID).Str("serviceId", template.ServiceID).Msg("duplicate provierId + serviceId detected")
		exitVal = 1
	}
	collision[template.ProviderID+"/"+template.ServiceID] = true

	Validator := validator.New()
	err = Validator.Struct(template)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			tlog.Warn().Err(err).Msg("template field validation")
			exitVal = 1
		}
	}

	if template.Version < 0 {
		tlog.Info().Msg("use of negative version number is not recommended")
	}

	if template.Shared {
		tlog.Info().Msg("shared flag has been deprecated, use sharedProviderName instead")
		// Override to ensure settings in pretty-print output are correct
		template.Shared = false
		template.SharedProviderName = true
	}

	if err := isUnreachable(template.Logo); err != nil {
		tlog.Warn().Err(err).Str("logoUrl", template.Logo).Msg("logo check failed")
	}

	if *Cloudflare {
		CloudflareTemplateChecks(template)
	}

	conflictingTypes := make(map[string]string)
	for rnum, record := range template.Records {
		checkRecord(rnum, record, conflictingTypes)
	}

	if *PrettyPrint || *Inplace {
		pp, err := json.Marshal(template)
		if err != nil {
			tlog.Error().Err(err).Msg("json marshaling failed")
		}
		var out bytes.Buffer
		json.Indent(&out, pp, "", "    ")
		fmt.Fprintf(&out, "\n")
		if *Inplace {
			writeBack(templatePath, out)
		} else {
			out.WriteTo(os.Stdout)
		}
	}
}

const validChars = "-.0123456789_abcdefghijklmnopqrstuvwxy"

func checkInvalidChars(s string) bool {
	for _, char := range s {
		if !strings.Contains(validChars, strings.ToLower(string(char))) {
			return true
		}
	}
	return false
}

func isUnreachable(logoUrl string) error {
	if !*CheckLogos || logoUrl == "" {
		return nil
	}
	resp, err := http.Get(logoUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status %d", resp.StatusCode)
	}
	return nil
}

func writeBack(templatePath string, out bytes.Buffer) {
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

func checkRecord(rnum int, record Record, conflictingTypes map[string]string) {
	rlog := tlog.With().Int("record", rnum).Logger()
	rlog.Debug().Str("type", record.Type).Str("groupid", record.GroupID).Str("host", record.Host).Msg("check record")

	if t, ok := conflictingTypes[record.GroupID+"/"+record.Host]; ok && (t == "CNAME" || record.Type == "CNAME") {
		rlog.Error().
			Str("groupid", record.GroupID).
			Str("host", record.Host).
			Str("type", record.Type).
			Str("othertype", t).
			Msg("CNAME cannot be mixed with other record types")
		exitVal = 1
	}
	conflictingTypes[record.GroupID+"/"+record.Host] = record.Type

	switch record.Type {
	case "CNAME", "NS":
		if record.Host == "@" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be @")
			exitVal = 1
		}
		fallthrough
	case "A", "AAAA":
		if record.Host == "" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be empty")
			exitVal = 1
		}

	case "TXT":
		if record.Host == "" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be empty")
			exitVal = 1
		}
		if record.Data == "" {
			rlog.Error().Str("type", record.Type).Msg("record data must not be empty")
			exitVal = 1
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
				exitVal = 1
			}
		}

	case "MX":
		if record.Host == "" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be empty")
			exitVal = 1
		}
		if record.PointsTo == "" {
			rlog.Error().Str("type", record.Type).Msg("record pointsTo must not be empty")
			exitVal = 1
		}
		if record.Priority < 0 || Max31b < record.Priority {
			rlog.Error().Str("type", record.Type).Int("priority", int(record.Priority)).Msg("invalid priority")
			exitVal = 1
		}

	case "SRV":
		if record.Name == "" {
			rlog.Error().Str("type", record.Type).Msg("record name must not be empty")
			exitVal = 1
		}
		if record.Target == "" {
			rlog.Error().Str("type", record.Type).Msg("record target must not be empty")
			exitVal = 1
		}
		if isInvalidProtocol(record.Protocol) {
			rlog.Warn().Str("type", record.Type).Str("protocol", record.Protocol).Msg("invalid protocol")
			exitVal = 1
		}
		if record.Priority < 0 || Max31b < record.Priority {
			rlog.Error().Str("type", record.Type).Int("priority", int(record.Priority)).Msg("invalid priority")
			exitVal = 1
		}
		if record.Service == "" {
			rlog.Error().Str("type", record.Type).Msg("record service must not be empty")
			exitVal = 1
		}
		if record.Weight < 0 || Max31b < record.Weight {
			rlog.Error().Str("type", record.Type).Int("weigth", int(record.Weight)).Msg("invalid weigth")
			exitVal = 1
		}
		if record.Port < 1 || Max16b < record.Port {
			rlog.Error().Str("type", record.Type).Int("port", int(record.Port)).Msg("invalid port")
			exitVal = 1
		}

	case "SPFM":
		if record.Host == "" {
			rlog.Error().Str("type", record.Type).Msg("record host must not be empty")
			exitVal = 1
		}
		if record.SPFRules == "" {
			rlog.Error().Str("type", record.Type).Msg("record spfRules must not be empty")
			exitVal = 1
		}

	case "APEXCNAME":
		if *Cloudflare {
			rlog.Info().Msg("Cloudflare does not support APEXCNAME, use CNAME instead")
		}

	case "REDIR301", "REDIR302":
		if record.Target == "" {
			rlog.Error().Str("type", record.Type).Msg("record target must not be empty")
			exitVal = 1
		}
	default:
		rlog.Info().Str("type", record.Type).Msg("unusual record type check DNS providers if they support it")
	}

	if isVariable(record.Type) {
		rlog.Error().Msg("record type must not be variable")
		exitVal = 1
	}

	if record.TTL < 0 || Max31b < record.TTL {
		rlog.Error().Str("type", record.Type).Int("ttl", int(record.TTL)).Msg("invalid TTL")
		exitVal = 1
	}

	if isVariable(record.GroupID) {
		rlog.Error().Msg("record groupId must not be variable")
		exitVal = 1
	}

	if isVariable(record.TxtCMP) {
		rlog.Error().Msg("record txtConflictMatchingPrefix must not be variable")
		exitVal = 1
	}

	if *Cloudflare {
		if record.Essential != "" {
			rlog.Info().Msg("Cloudflare does not support essential record settings")
		}
	}
}

func isInvalidProtocol(proto string) bool {
	switch strings.ToLower(proto) {
	case "_tcp", "_udp", "_tls":
		return false
	}
	return true
}

func isVariable(s string) bool {
	return strings.Count(s, "%") > 1
}

func CloudflareTemplateChecks(template Template) {
	if template.SyncBlock {
		tlog.Error().Msg("Cloudflare does not support syncBlock")
		exitVal = 1
	}
	if template.SyncPubKeyDomain == "" {
		tlog.Error().Msg("Cloudflare requires syncPubKeyDomain")
		exitVal = 1
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
}
