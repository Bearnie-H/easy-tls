// Command rule-editor implements a basic command-line utility for managing an EasyTLS Proxy Rules file.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Bearnie-H/easy-tls/proxy"
)

// Define the command-line flags
var (
	AddRulesFlag    = flag.Bool("add", false, "Flag indicating that you want to add new rules to the given EasyTLS Proxy Rules file.")
	DeleteRulesFlag = flag.Bool("delete", false, "Flag indicating that you want to remove existing rules from the given EasyTLS Proxy Rules file.")
	EditRulesFlag   = flag.Bool("edit", false, "Flag indicating whether you want to simply edit existing rules from the given EasyTLS Proxy Rules file.")
	RulesFilename   = flag.String("file", "EasyTLS-Proxy.rules", "The filename of the EasyTLS Proxy Rules file to work with.")
)

var (
	rules proxy.ReverseProxyRuleSet = proxy.ReverseProxyRuleSet{}
)

func main() {
	flag.Parse()

	fmt.Printf("Working with rules file: %s.\n\n", *RulesFilename)

	if err := DecodeFile(*RulesFilename, &rules); err != nil {
		fmt.Printf("Error: Failed to decode EasyTLS Proxy Rule file %s - %s.\n", *RulesFilename, err)
		os.Exit(1)
	}

	ListRules(rules)

	if *EditRulesFlag {
		*AddRulesFlag = true
		*DeleteRulesFlag = true
	}

	if *AddRulesFlag {
		rules = AddRules(rules)
		ListRules(rules)
	}

	if *DeleteRulesFlag {
		rules = DeleteRules(rules)
		ListRules(rules)
	}

	if *AddRulesFlag || *DeleteRulesFlag {
		if err := EncodeFile(*RulesFilename, rules); err != nil {
			fmt.Printf("Error: Failed to save EasyTLS Proxy Rule file %s - %s.\n", *RulesFilename, err)
			os.Exit(1)
		}
		fmt.Printf("All changes saved to %s.\n", *RulesFilename)
	}
}

// DecodeFile will read and parse the given Rules file a manageable set of rules to work with.
func DecodeFile(Filename string, rules *proxy.ReverseProxyRuleSet) error {

	// Open the file for reading
	f, err := os.Open(Filename)
	if os.IsNotExist(err) {
		return EncodeFile(Filename, *rules)
	}
	if err != nil {
		return err
	}
	defer f.Close()

	// Decode the JSON-formatted rules into an array of Rules
	return json.NewDecoder(f).Decode(rules)
}

// EncodeFile will JSON encode the Proxy rules, writing them back to the original file.
func EncodeFile(Filename string, rules proxy.ReverseProxyRuleSet) error {

	// Create (or truncate) the Rules file.
	f, err := os.Create(Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	sort.Slice(rules, rules.Less)

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "\t")
	return encoder.Encode(rules)
}

// AddRules will allow the user to add new rules to the rules file.
func AddRules(RuleSet []proxy.ReverseProxyRoutingRule) []proxy.ReverseProxyRoutingRule {
	next := true

	fmt.Print("Adding new routing rules to the rules file...\n")

	for next {
		fmt.Print("\n")
		if NewRule, err := GetRuleToAdd(); err == nil {
			RuleSet = append(RuleSet, *NewRule)
		} else {
			fmt.Printf("Error: Invalid Rule ignored - %s.\n", err)
		}
		next = Continue()
	}

	fmt.Print("Finished adding new routing rules to the rules file!\n\n")

	return RuleSet
}

// GetRuleToAdd will prompt the user to entry the necessary fields for adding a new rule.
func GetRuleToAdd() (*proxy.ReverseProxyRoutingRule, error) {
	NewRule := &proxy.ReverseProxyRoutingRule{}

	NewRule.PathPrefix = GetString("Enter the URI prefix the rule should match on: ")
	NewRule.DestinationHost = GetString("Enter the Destination Host this rule should forward to: ")
	if temp, err := GetInt("Enter the Destination Port this rule should forward to: "); err == nil {
		NewRule.DestinationPort = temp
	} else {
		return nil, err
	}
	if strings.ToLower(GetString("Should this rule strip the prefix when forwarding? [y/N]: ")) == "y" {
		NewRule.StripPrefix = true
	} else {
		NewRule.StripPrefix = false
	}

	if !strings.HasPrefix(NewRule.PathPrefix, "/") {
		NewRule.PathPrefix = "/" + NewRule.PathPrefix
	}

	if strings.HasSuffix(NewRule.PathPrefix, "/") {
		NewRule.PathPrefix = strings.TrimSuffix(NewRule.PathPrefix, "/")
	}

	if NewRule.DestinationPort < 0 || NewRule.DestinationPort > (256*256) {
		return nil, fmt.Errorf("easytls proxy error - Invalid Destination Port (%d) - Out of range", NewRule.DestinationPort)
	}

	fmt.Printf("Adding new rule (%s).\n", NewRule.String())

	return NewRule, nil
}

// DeleteRules will allow the user to delete existing rules to the rules file.
func DeleteRules(RuleSet []proxy.ReverseProxyRoutingRule) []proxy.ReverseProxyRoutingRule {
	newRules := []proxy.ReverseProxyRoutingRule{}

	fmt.Print("Deleting existing routing rules from the rules file...\n\n")

	for _, Rule := range RuleSet {
		if strings.ToLower(GetString(fmt.Sprintf("Do you want to delete routing rule: (%s)? [y/N]: ", Rule.String()))) != "y" {
			newRules = append(newRules, Rule)
		}
	}

	fmt.Print("Finished deleting existing routing rules from the rules file!\n\n")

	return newRules
}

// ListRules will allow simply print the existing rules to the user in a consistent format.
func ListRules(RuleSet []proxy.ReverseProxyRoutingRule) {

	if len(RuleSet) == 0 {
		fmt.Println("There are no rules to list.")
		return
	}

	sort.Slice(rules, rules.Less)

	fmt.Println("Current proxy rules:")
	for _, Rule := range RuleSet {
		fmt.Println("\t" + Rule.String())
	}
	fmt.Print("\n")
}

// GetString will prompt the user for a value, and return the entered string.
func GetString(prompt string) string {
	input := ""
	fmt.Print(prompt)
	fmt.Scanln(&input)
	return input
}

// GetInt will prompt the user for a value, and return the entered number.
func GetInt(prompt string) (int, error) {
	input := ""
	fmt.Print(prompt)
	fmt.Scanln(&input)
	return strconv.Atoi(input)
}

// Continue will check if the user wants to continue adding or deleting.
func Continue() bool {
	return strings.ToLower(GetString("Do you want to continue? [y/N]: ")) == "y"
}
