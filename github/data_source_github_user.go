package github

import (
	"context"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubUserRead,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "username"},
			},
			"login": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gravatar_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"site_admin": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"company": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"blog": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bio": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gpg_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ssh_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_repos": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"public_gists": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"followers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"following": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"suspended_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ExactlyOneOf: []string{"id", "username"},
			},
		},
	}
}

func dataSourceGithubUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	var user *github.User
	var err error

	if d.Get("username") != "" {
		username := d.Get("username").(string)

		user, _, err = client.Users.Get(ctx, username)
		if err != nil {
			return err
		}
	} else if d.Get("id") != 0 {
		id := int64(d.Get("id").(int))

		user, _, err = client.Users.GetByID(ctx, id)
		if err != nil {
			return err
		}
	}

	gpg, _, err := client.Users.ListGPGKeys(ctx, user.GetLogin(), nil)
	if err != nil {
		return err
	}
	ssh, _, err := client.Users.ListKeys(ctx, user.GetLogin(), nil)
	if err != nil {
		return err
	}

	gpgKeys := []string{}
	for _, v := range gpg {
		gpgKeys = append(gpgKeys, v.GetPublicKey())
	}

	sshKeys := []string{}
	for _, v := range ssh {
		sshKeys = append(sshKeys, v.GetKey())
	}

	d.SetId(strconv.FormatInt(user.GetID(), 10))
	d.Set("username", user.GetLogin())
	if err = d.Set("login", user.GetLogin()); err != nil {
		return err
	}
	if err = d.Set("avatar_url", user.GetAvatarURL()); err != nil {
		return err
	}
	if err = d.Set("gravatar_id", user.GetGravatarID()); err != nil {
		return err
	}
	if err = d.Set("site_admin", user.GetSiteAdmin()); err != nil {
		return err
	}
	if err = d.Set("company", user.GetCompany()); err != nil {
		return err
	}
	if err = d.Set("blog", user.GetBlog()); err != nil {
		return err
	}
	if err = d.Set("location", user.GetLocation()); err != nil {
		return err
	}
	if err = d.Set("name", user.GetName()); err != nil {
		return err
	}
	if err = d.Set("email", user.GetEmail()); err != nil {
		return err
	}
	if err = d.Set("bio", user.GetBio()); err != nil {
		return err
	}
	if err = d.Set("gpg_keys", gpgKeys); err != nil {
		return err
	}
	if err = d.Set("ssh_keys", sshKeys); err != nil {
		return err
	}
	if err = d.Set("public_repos", user.GetPublicRepos()); err != nil {
		return err
	}
	if err = d.Set("public_gists", user.GetPublicGists()); err != nil {
		return err
	}
	if err = d.Set("followers", user.GetFollowers()); err != nil {
		return err
	}
	if err = d.Set("following", user.GetFollowing()); err != nil {
		return err
	}
	if err = d.Set("created_at", user.GetCreatedAt().String()); err != nil {
		return err
	}
	if err = d.Set("updated_at", user.GetUpdatedAt().String()); err != nil {
		return err
	}
	if err = d.Set("suspended_at", user.GetSuspendedAt().String()); err != nil {
		return err
	}
	if err = d.Set("node_id", user.GetNodeID()); err != nil {
		return err
	}

	return nil
}
