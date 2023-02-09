package jupiterone

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jupiterone/terraform-provider-jupiterone/jupiterone/internal/client"
)

var createAlertActionJSON = `{"type":"CREATE_ALERT"}`
var testRuleResourceName = "jupiterone_rule.test"

func TestInlineRuleInstance_Basic(t *testing.T) {
	ctx := context.TODO()

	recordingClient, directClient, cleanup := setupTestClients(ctx, t)
	defer cleanup(t)

	ruleName := "tf-provider-test-rule"
	operations := getValidOperations()
	operationsUpdate := getValidOperationsWithoutFilter()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(recordingClient),
		CheckDestroy:             testAccCheckRuleInstanceDestroy(ctx, directClient),
		Steps: []resource.TestStep{
			{
				Config: testInlineRuleInstanceBasicConfigWithOperations(ruleName, operations),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, testRuleResourceName, directClient),
					resource.TestCheckResourceAttrSet(testRuleResourceName, "id"),
					resource.TestCheckResourceAttr(testRuleResourceName, "version", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "name", ruleName),
					resource.TestCheckResourceAttr(testRuleResourceName, "description", "Test"),
					resource.TestCheckResourceAttr(testRuleResourceName, "spec_version", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "polling_interval", "ONE_WEEK"),
					resource.TestCheckResourceAttr(testRuleResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "tags.0", "tf_acc:1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "tags.1", "tf_acc:2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "operations.#", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "operations.0.actions.#", "2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "operations.0.actions.1", createAlertActionJSON),
					resource.TestCheckResourceAttr(testRuleResourceName, "outputs.#", "2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "outputs.0", "queries.query0.total"),
					resource.TestCheckResourceAttr(testRuleResourceName, "outputs.1", "alertLevel"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.#", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.0.queries.#", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.0.queries.0.name", "query0"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.0.queries.0.version", "v1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.0.queries.0.query", "Find DataStore with classification=('critical' or 'sensitive' or 'confidential' or 'restricted') and encrypted!=true"),
				),
			},
			{
				Config: testInlineRuleInstanceBasicConfigWithOperations(ruleName, operationsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, testRuleResourceName, directClient),
					resource.TestCheckResourceAttrSet(testRuleResourceName, "id"),
					resource.TestCheckResourceAttr(testRuleResourceName, "version", "2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "name", ruleName),
					resource.TestCheckResourceAttr(testRuleResourceName, "description", "Test"),
					resource.TestCheckResourceAttr(testRuleResourceName, "spec_version", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "polling_interval", "ONE_WEEK"),
					resource.TestCheckResourceAttr(testRuleResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "tags.0", "tf_acc:1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "tags.1", "tf_acc:2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "operations.0.actions.1", createAlertActionJSON),
					resource.TestCheckResourceAttr(testRuleResourceName, "operations.#", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "operations.0.actions.#", "2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "outputs.#", "2"),
					resource.TestCheckResourceAttr(testRuleResourceName, "outputs.0", "queries.query0.total"),
					resource.TestCheckResourceAttr(testRuleResourceName, "outputs.1", "alertLevel"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.#", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.0.queries.#", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.0.queries.0.name", "query0"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.0.queries.0.version", "v1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.0.queries.0.query", "Find DataStore with classification=('critical' or 'sensitive' or 'confidential' or 'restricted') and encrypted!=true"),
				),
			},
		},
	})
}

func TestReferencedQuestionRule_Basic(t *testing.T) {
	ctx := context.TODO()

	recordingClient, directClient, cleanup := setupTestClients(ctx, t)
	defer cleanup(t)

	ruleName := "tf-provider-test-rule"
	operations := getValidOperations()
	operationsUpdate := getValidOperationsWithoutFilter()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(recordingClient),
		CheckDestroy:             testAccCheckRuleInstanceDestroy(ctx, directClient),
		Steps: []resource.TestStep{
			{
				Config: testReferencedRuleInstanceBasicConfigWithOperations(ruleName, operations),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, testRuleResourceName, directClient),
					resource.TestCheckResourceAttrSet(testRuleResourceName, "id"),
					resource.TestCheckResourceAttr(testRuleResourceName, "version", "1"),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.#", "0"),
					resource.TestCheckResourceAttrPair("jupiterone_question.test", "id", testRuleResourceName, "question_id"),
				),
			},
			{
				Config: testReferencedRuleInstanceBasicConfigWithOperations(ruleName, operationsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, testRuleResourceName, directClient),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.#", "0"),
					resource.TestCheckResourceAttrPair("jupiterone_question.test", "id", testRuleResourceName, "question_id"),
				),
			},
			{
				Config: testInlineRuleInstanceBasicConfigWithOperations(ruleName, operationsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, testRuleResourceName, directClient),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.#", "1"),
					resource.TestCheckNoResourceAttr(testRuleResourceName, "question_id"),
				),
			},
			{
				Config: testReferencedRuleInstanceBasicConfigWithOperations(ruleName, operationsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, testRuleResourceName, directClient),
					resource.TestCheckResourceAttr(testRuleResourceName, "question.#", "0"),
					resource.TestCheckResourceAttrPair("jupiterone_question.test", "id", testRuleResourceName, "question_id"),
				),
			},
		},
	})
}

func TestRuleInstance_Config_Errors(t *testing.T) {
	ctx := context.TODO()

	recordingClient, _, cleanup := setupTestClients(ctx, t)
	defer cleanup(t)

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(recordingClient),
		Steps: []resource.TestStep{
			{
				Config:      testInlineRuleInstanceBasicConfigWithOperations(rName, "\"not json\""),
				ExpectError: regexp.MustCompile(`list of object required`),
			},
			{
				Config:      testInlineRuleInstanceBasicConfigWithOperations(rName, getInvalidOperations()),
				ExpectError: regexp.MustCompile(`string value must be valid JSON`),
			},
			{
				Config:      testInlineRuleInstanceBasicConfigWithOperations("", getValidOperations()),
				ExpectError: regexp.MustCompile(`Attribute name string length must be between 1 and 255, got: 0`),
			},
			{
				Config:      testRuleInstanceBasicConfigWithPollingInterval(rName, "INVALID_POLLING_INTERVAL"),
				ExpectError: regexp.MustCompile(`Attribute polling_interval value must be one of:`),
			},
		},
	})
}

func testAccCheckRuleExists(ctx context.Context, ruleName string, qlient graphql.Client) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource := s.RootModule().Resources[ruleName]

		return ruleExistsHelper(ctx, resource.Primary.ID, qlient)
	}
}

func ruleExistsHelper(ctx context.Context, id string, qlient graphql.Client) error {
	if qlient == nil {
		return nil
	}

	duration := 10 * time.Second
	err := resource.RetryContext(ctx, duration, func() *resource.RetryError {
		_, err := client.GetQuestionRuleInstance(ctx, qlient, id)

		if err == nil {
			return nil
		}

		if err != nil && strings.Contains(err.Error(), "Rule instance does not exist.") {
			return resource.RetryableError(fmt.Errorf("Rule instance does not exist (id=%q)", id))
		}

		return resource.NonRetryableError(err)
	})

	if err != nil {
		return err
	}

	return nil
}

func testAccCheckRuleInstanceDestroy(ctx context.Context, qlient graphql.Client) func(*terraform.State) error {
	return func(s *terraform.State) error {
		resource := s.RootModule().Resources[testRuleResourceName]

		return ruleInstanceDestroyHelper(ctx, resource.Primary.ID, qlient)
	}
}

func ruleInstanceDestroyHelper(ctx context.Context, id string, qlient graphql.Client) error {
	if qlient == nil {
		return nil
	}

	duration := 10 * time.Second
	err := resource.RetryContext(ctx, duration, func() *resource.RetryError {
		_, err := client.GetQuestionRuleInstance(ctx, qlient, id)

		if err == nil {
			return resource.RetryableError(fmt.Errorf("Rule instance still exists (id=%q)", id))
		}

		if err != nil && strings.Contains(err.Error(), "Rule instance does not exist.") {
			return nil
		}

		return resource.NonRetryableError(err)
	})

	if err != nil {
		return err
	}

	return nil
}

func getInvalidOperations() string {
	return `[
		{
			when = "not json"
			actions = [
					"still not json",
					"also not json",
				]
			}
	]`
}

func getValidOperations() string {
	return fmt.Sprintf(`[
			{
				when = %q
				actions = [
					%q,
					%q,
				]
			}
		]`,
		`{"type":"FILTER","specVersion":1,"condition":"{{queries.query0.total != 0}}"}`,
		`{"targetValue":"HIGH","type":"SET_PROPERTY","targetProperty":"alertLevel"}`,
		createAlertActionJSON)
}

func getValidOperationsWithoutFilter() string {
	return fmt.Sprintf(`[
			{
				actions = [
					%q,
					%q,
				]
			}
		]`,
		`{"targetValue":"HIGH","type":"SET_PROPERTY","targetProperty":"alertLevel"}`,
		createAlertActionJSON)
}

func testInlineRuleInstanceBasicConfigWithOperations(rName string, operations string) string {
	return fmt.Sprintf(`
		resource "jupiterone_rule" "test" {
			name = %q
			description = "Test"
			spec_version = 1
			polling_interval = "ONE_WEEK"
			tags = ["tf_acc:1","tf_acc:2"]

			question {
				queries {
					name = "query0"
					query = "Find DataStore with classification=('critical' or 'sensitive' or 'confidential' or 'restricted') and encrypted!=true"
					version = "v1"
				}
			}

			outputs = [
				"queries.query0.total",
				"alertLevel"
			]

			operations = %s
		}
	`, rName, operations)
}

func testReferencedRuleInstanceBasicConfigWithOperations(rName string, operations string) string {
	return fmt.Sprintf(`
		resource "jupiterone_question" "test" {
			title = %q
			description = "Test"
			tags = ["tf_acc:1","tf_acc:2"]

			query {
				name = "query0"
				query = "Find DataStore with classification=('critical' or 'sensitive' or 'confidential' or 'restricted') and encrypted!=true"
				version = "v1"
			}
		}

		resource "jupiterone_rule" "test" {
			name = %q
			description = "Test"
			spec_version = 1
			polling_interval = "ONE_WEEK"
			tags = ["tf_acc:1","tf_acc:2"]

			question_id = jupiterone_question.test.id

			outputs = [
				"queries.query0.total",
				"alertLevel"
			]

			operations = %s
		}
	`, rName, rName, operations)
}

func testRuleInstanceBasicConfigWithPollingInterval(rName string, pollingInterval string) string {
	return fmt.Sprintf(`
		provider "jupiterone" {}

		resource "jupiterone_rule" "test" {
			name = %q
			description = "Test"
			spec_version = 1
			polling_interval = %q

			tags = ["tf_acc:1","tf_acc:2"]
			question {
				queries {
					name = "query0"
					query = "Find DataStore with classification=('critical' or 'sensitive' or 'confidential' or 'restricted') and encrypted!=true"
					version = "v1"
				}
			}

			outputs = [
				"queries.query0.total",
				"alertLevel"
			]

			operations = %s
		}
	`, rName, pollingInterval, getValidOperations())
}
