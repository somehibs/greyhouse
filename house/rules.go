package house

import (
	"log"
	"errors"
	"io/ioutil"
	"encoding/json"
	"time"

	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
)

type RuleList map[string]*api.Rule
const ruleConfigFile = "rules.json"

type RuleService struct {
	// Store all the rules here, based on their rule key.
	rules RuleList
	// Store the rules based on the Room they affect.
	appliesToRoom map[api.Room][]api.Rule
	// returned for every room
	global map[string]*api.Rule
}

func NewRuleService() RuleService {
	log.Print("Starting rule service...")
	service := RuleService{RuleList{}, make(map[api.Room][]api.Rule, 0), map[string]*api.Rule{}}
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
		log.Panicf("Could not unmarshal list because %+v - used bytes %s", err, in)
	}
	for _, rule := range ruleList {
		log.Printf("Loaded file rule %+v", rule)
		rs.Create(nil, rule)
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
	isGlobal := false
	for _, condition := range rule.Conditions {
		if condition.Room == 0 || (condition.TimeStart == 0 && condition.TimeEnd == 0) {
			isGlobal = true
			continue
		}
		if rs.appliesToRoom[condition.Room] != nil {
			rs.appliesToRoom[condition.Room] = append(rs.appliesToRoom[condition.Room], *rule)
		} else {
			rs.appliesToRoom[condition.Room] = []api.Rule{*rule}
		}
	}
	if len(rule.Conditions) == 0 {
		log.Printf("no conditions set, letting rule \"%s\" go global", rule.Name)
		isGlobal = true
	}
	if isGlobal {
		log.Printf("Adding global rule: %s", rule.Name)
		rs.global[rule.Name] = rule
	}
	return &api.CreateRuleResponse{}, nil
}

func (rs RuleService) Delete(ctx context.Context, toDelete *api.Rule) (*api.DeleteRuleResponse, error) {
	if rs.rules[toDelete.Name] == nil {
		return nil, errors.New("not found")
	}
	rule := rs.rules[toDelete.Name]
	for _, modifier := range rule.Conditions {
		if rs.appliesToRoom[modifier.Room] != nil {
			applyList := rs.appliesToRoom[modifier.Room]
			deleteRules := make([]int, 0)
			for index, modifierRule := range applyList {
				if modifierRule.Name == toDelete.Name {
					deleteRules = append(deleteRules, index)
				}
			}
			for i := len(deleteRules)-1; i >= 0; i-- {
				applyList = append(applyList[:deleteRules[i]], applyList[:deleteRules[i+1]]...)
			}
			rs.appliesToRoom[modifier.Room] = applyList
		}
	}
	delete(rs.rules, toDelete.Name)
	if rs.global[toDelete.Name] != nil {
		delete(rs.global, toDelete.Name)
	}
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

func (rs RuleService) ApplyRules(room api.Room) []*api.RuleEffect {
	rules := make([]*api.RuleEffect, 0)
	for _, v := range rs.global {
		if checkConditions(room, v.Conditions) {
			rules = append(rules, v.Modifiers...)
		}
	}
	for _, v := range rs.appliesToRoom[room] {
		if checkConditions(room, v.Conditions) {
			rules = append(rules, v.Modifiers...)
		}
	}
	return rules
}

func getMidnightToday() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Now().Location())
}

func checkConditions(room api.Room, conditions []*api.RuleConditions) bool {
	log.Printf("Checking conditions: %+v", conditions)
	for _, condition := range conditions {
		cond := *condition
		if inRange(cond) && roomMatch(cond, room) {
			return true
		}
	}
	return false
}

func roomMatch(condition api.RuleConditions, room api.Room) bool {
	if condition.Room == 0 {
		return true
	}
	if condition.Room == room {
		return true
	}
	return false
}

func inRange(condition api.RuleConditions) bool {
	if condition.TimeStart == 0 && condition.TimeEnd == 0 {
		return true
	}
	midnight := getMidnightToday()
	start := midnight.Add(time.Duration(condition.TimeStart)*time.Second)
	end := midnight.Add(time.Duration(condition.TimeEnd)*time.Second)
	now := time.Now()
	match := now.After(start) && now.Before(end)
	//log.Printf("Start %s End %s", start, end)
	//log.Printf("range match %s", match)
	return match
}
