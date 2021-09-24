package github

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubProjectCard() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubProjectCardCreate,
		Read:   resourceGithubProjectCardRead,
		Update: resourceGithubProjectCardUpdate,
		Delete: resourceGithubProjectCardDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubProjectCardImport,
		},
		Schema: map[string]*schema.Schema{
			"column_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"note": {
				Type:     schema.TypeString,
				Required: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"card_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceGithubProjectCardCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	columnIDStr := d.Get("column_id").(string)
	columnID, err := strconv.ParseInt(columnIDStr, 10, 64)
	if err != nil {
		return unconvertibleIdErr(columnIDStr, err)
	}

	log.Printf("[DEBUG] Creating project card note in column ID: %d", columnID)
	client := meta.(*Owner).v3client
	options := github.ProjectCardOptions{Note: d.Get("note").(string)}
	ctx := context.Background()
	card, _, err := client.Projects.CreateProjectCard(ctx, columnID, &options)
	if err != nil {
		return err
	}

	d.Set("card_id", card.GetID())
	d.SetId(card.GetNodeID())

	return resourceGithubProjectCardRead(d, meta)
}

func resourceGithubProjectCardRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	nodeID := d.Id()
	cardID := d.Get("card_id").(int)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	log.Printf("[DEBUG] Reading project card: %s", nodeID)
	card, _, err := client.Projects.GetProjectCard(ctx, int64(cardID))
	if err != nil {
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing project card %s from state because it no longer exists in GitHub", d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	// FIXME: Remove URL parsing if a better option becomes available
	columnURL := card.GetColumnURL()
	columnIDStr := strings.TrimPrefix(columnURL, client.BaseURL.String()+`projects/columns/`)
	if err != nil {
		return unconvertibleIdErr(columnIDStr, err)
	}

	d.Set("note", card.GetNote())
	d.Set("column_id", columnIDStr)
	d.Set("card_id", card.GetID())

	return nil
}

func resourceGithubProjectCardUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	cardID := d.Get("card_id").(int)

	log.Printf("[DEBUG] Updating project Card: %s", d.Id())
	options := github.ProjectCardOptions{
		Note: d.Get("note").(string),
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	_, _, err := client.Projects.UpdateProjectCard(ctx, int64(cardID), &options)
	if err != nil {
		return err
	}

	return resourceGithubProjectCardRead(d, meta)
}

func resourceGithubProjectCardDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting project Card: %s", d.Id())
	cardID := d.Get("card_id").(int)
	_, err := client.Projects.DeleteProjectCard(ctx, int64(cardID))
	if err != nil {
		return err
	}

	return nil
}

func resourceGithubProjectCardImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	cardIDStr := d.Id()
	cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
	if err != nil {
		return []*schema.ResourceData{d}, unconvertibleIdErr(cardIDStr, err)
	}

	log.Printf("[DEBUG] Importing project card with card ID: %d", cardID)
	client := meta.(*Owner).v3client
	ctx := context.Background()
	card, _, err := client.Projects.GetProjectCard(ctx, cardID)
	if card == nil || err != nil {
		return []*schema.ResourceData{d}, err
	}

	d.SetId(card.GetNodeID())
	d.Set("card_id", cardID)

	return []*schema.ResourceData{d}, nil

}
