package house

import (
	"log"
	"errors"
	"io/ioutil"
	"encoding/json"

	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
)

type RuleList map[string]*api.Rule
const ruleConfigFile = "rules.json"

type RuleService struct {
	// Store all the rules here, based on their rule key.
	rules RuleList
	// Store the rules based on the Room they affect, allowing for cheaper 'can i do this' queries
	appliesTo map[api.Room][]api.Rule
}

func NewRuleService() RuleService {
	log.Print("Starting rule service...")
	service := RuleService{RuleList{}, make(map[api.Room][]api.Rule, 0)}
	service.ReadRules()
	return service
}

func (rs RuleService) ReadRules() {
	in, err := ioutil.ReadFile(ruleConfigFile)
	if err != nil {
		log.Printf("Could not read rules.json: %+v", err)
		return
	}
	ruleList := RuleList{}
	err = json.Unmarshal(in, &ruleList)
	if err != nil {
		log.Printf("Could not unmarshal list because %+v - used bytes %s", err, in)
	}
}

func (rs RuleService) WriteRules() {
	out, err := json.Marshal(rs.rules)
	if err != nil {
		log.Printf("Error marshalling json %+v", err)
		return
	}
	err = ioutil.WriteFile(ruleConfigFile, out, 0644)
	if err != nil {
		log.Printf("Error writing config %+v", err)
	}
}

func (rs RuleService) Create(ctx context.Context, rule *api.Rule) (*api.CreateRuleResponse, error) {
	if rs.rules[rule.Name] != nil {
		return nil, errors.New("already_exists")
	}
	// Create the rule, put it in the right places.
	rs.rules[rule.Name] = rule
	for _, modifier := range rule.Modifiers {
		if rs.appliesTo[modifier.Room] != nil {
			rs.appliesTo[modifier.Room] = append(rs.appliesTo[modifier.Room], *rule)
		} else {
			rs.appliesTo[modifier.Room] = []api.Rule{*rule}
		}
	}
	return &api.CreateRuleResponse{}, nil
}

func (rs RuleService) Delete(ctx context.Context, toDelete *api.Rule) (*api.DeleteRuleResponse, error) {
	if rs.rules[toDelete.Name] == nil {
		return nil, errors.New("not found")
	}
	rule := rs.rules[toDelete.Name]
	for _, modifier := range rule.Modifiers {
		if rs.appliesTo[modifier.Room] != nil {
			applyList := rs.appliesTo[modifier.Room]
			deleteRules := make([]int, 0)
			for index, modifierRule := range applyList {
				if modifierRule.Name == toDelete.Name {
					deleteRules = append(deleteRules, index)
				}
			}
			for i := len(deleteRules)-1; i >= 0; i-- {
				applyList = append(applyList[:deleteRules[i]], applyList[:deleteRules[i+1]]...)
			}
			rs.appliesTo[modifier.Room] = applyList
		}
	}
	delete(rs.rules, toDelete.Name)
	return &api.DeleteRuleResponse{}, nil
}

func (rs RuleService) List(ctx context.Context, filter *api.RuleFilter) (*api.RuleList, error) {
	ruleList := make([]*api.Rule, len(rs.rules))
	var i = 0
	for k := range rs.rules {
		ruleList[i] = rs.rules[k]
		i += 1
	}
	return &api.RuleList{Rules: ruleList}, nil
}
