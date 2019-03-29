package rancher2

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	//managementClient "github.com/rancher/types/client/management/v3"
)

const (
	bootstrapDefaultTokenDesc = "Terraform bootstrap admin token"
	bootstrapDefaultUser      = "admin"
	bootstrapDefaultPassword  = "admin"
	bootstrapDefaultTTL       = "60000"
	bootstrapSettingURL       = "server-url"
	bootstrapSettingTelemetry = "telemetry-opt"
)

//Schemas

func bootstrapFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"current_password": &schema.Schema{
			Type:      schema.TypeString,
			Optional:  true,
			Computed:  true,
			Sensitive: true,
		},
		"password": &schema.Schema{
			Type:      schema.TypeString,
			Optional:  true,
			Computed:  true,
			Sensitive: true,
		},
		"token_ttl": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"token": &schema.Schema{
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},
		"token_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"token_update": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"telemetry": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"url": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"user": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
	}

	return s
}

func resourceRancher2Bootstrap() *schema.Resource {
	return &schema.Resource{
		Create: resourceRancher2BootstrapCreate,
		Read:   resourceRancher2BootstrapRead,
		Update: resourceRancher2BootstrapUpdate,
		Delete: resourceRancher2BootstrapDelete,
		Schema: bootstrapFields(),
	}
}

func resourceRancher2BootstrapCreate(d *schema.ResourceData, meta interface{}) error {
	if !meta.(*Config).Bootstrap {
		return fmt.Errorf("[ERROR] Resource rancher2_bootstrap just available on bootsrap mode")
	}

	err := bootstrapDoLogin(d, meta)
	if err != nil {
		return err
	}

	// Set rancher url
	url := strings.TrimSuffix(meta.(*Config).URL, "/v3")
	err = meta.(*Config).SetSetting(bootstrapSettingURL, url)
	if err != nil {
		return err
	}

	// Set telemetry option
	telemetry := "out"
	if d.Get("telemetry").(bool) {
		telemetry = "in"
	}

	err = meta.(*Config).SetSetting(bootstrapSettingTelemetry, telemetry)
	if err != nil {
		return err
	}

	// Generate a new token
	tokenID, token, err := meta.(*Config).GenerateUserToken(bootstrapDefaultUser, bootstrapDefaultTokenDesc, d.Get("token_ttl").(int))
	if err != nil {
		return fmt.Errorf("[ERROR] Creating Admin token: %s", err)
	}

	// Update new tokenkey
	d.Set("token_id", tokenID)
	d.Set("token", token)
	err = meta.(*Config).UpdateToken(token)
	if err != nil {
		return fmt.Errorf("[ERROR] Updating Admin token: %s", err)
	}

	// Set admin user password
	pass := d.Get("password").(string)
	newPass, adminUser, err := meta.(*Config).SetUserPasswordByName(bootstrapDefaultUser, pass)
	if err != nil {
		return fmt.Errorf("[ERROR] Updating Admin password: %s", err)
	}

	d.Set("password", newPass)
	d.Set("current_password", newPass)

	// Set resource ID
	d.SetId(adminUser.ID)

	return resourceRancher2BootstrapRead(d, meta)
}

func resourceRancher2BootstrapRead(d *schema.ResourceData, meta interface{}) error {
	if !meta.(*Config).Bootstrap {
		return fmt.Errorf("[ERROR] Resource rancher2_bootstrap just available on bootsrap mode")
	}

	err := bootstrapDoLogin(d, meta)
	if err != nil {
		return err
	}

	// Get rancher url
	url, err := meta.(*Config).GetSettingValue(bootstrapSettingURL)
	if err != nil {
		return err
	}

	d.Set("url", url)

	// Get telemetry
	telemetry, err := meta.(*Config).GetSettingValue(bootstrapSettingTelemetry)
	if err != nil {
		return err
	}

	if telemetry == "in" {
		d.Set("telemetry", true)
	} else {
		d.Set("telemetry", false)
	}

	return nil
}

func resourceRancher2BootstrapUpdate(d *schema.ResourceData, meta interface{}) error {
	err := bootstrapDoLogin(d, meta)
	if err != nil {
		return err
	}

	// Set rancher url
	url := strings.TrimSuffix(meta.(*Config).URL, "/v3")
	err = meta.(*Config).SetSetting(bootstrapSettingURL, url)
	if err != nil {
		return err
	}

	// Set telemetry option
	telemetry := "out"
	if d.Get("telemetry").(bool) {
		telemetry = "in"
	}

	err = meta.(*Config).SetSetting(bootstrapSettingTelemetry, telemetry)
	if err != nil {
		return err
	}

	// Regenerate a new token
	if d.Get("token_update").(bool) {
		tokenID, token, err := meta.(*Config).GenerateUserToken(bootstrapDefaultUser, bootstrapDefaultTokenDesc, d.Get("token_ttl").(int))
		if err != nil {
			return fmt.Errorf("[ERROR] Creating Admin token: %s", err)
		}

		// Delete old token
		err = meta.(*Config).DeleteToken(d.Get("token_id").(string))
		if err != nil {
			return fmt.Errorf("[ERROR] Deleting previous Admin token: %s", err)
		}

		// Update new tokenkey
		d.Set("token_id", tokenID)
		d.Set("token", token)
		err = meta.(*Config).UpdateToken(token)
		if err != nil {
			return fmt.Errorf("[ERROR] Updating Admin token: %s", err)
		}
	}

	// Set admin user password
	pass := d.Get("password").(string)
	newPass, adminUser, err := meta.(*Config).SetUserPasswordByName(bootstrapDefaultUser, pass)
	if err != nil {
		return fmt.Errorf("[ERROR] Updating Admin password: %s", err)
	}

	d.Set("password", newPass)
	d.Set("current_password", newPass)

	// Set resource ID
	d.SetId(adminUser.ID)

	return resourceRancher2BootstrapRead(d, meta)
}

func resourceRancher2BootstrapDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")

	return nil
}

func bootstrapDoLogin(d *schema.ResourceData, meta interface{}) error {
	// Try to connect with admin token
	token := d.Get("token").(string)
	err := meta.(*Config).UpdateToken(token)
	if err == nil {
		log.Printf("[INFO] Connecting with token")
		return nil
	}

	// If fails, try to login with default admin user and current password
	currentPass := d.Get("current_password").(string)
	if len(currentPass) == 0 {
		currentPass = bootstrapDefaultPassword
	}
	token, err = DoUserLogin(meta.(*Config).URL, bootstrapDefaultUser, currentPass, bootstrapDefaultTTL, meta.(*Config).CACerts, meta.(*Config).Insecure)
	if err != nil {
		return fmt.Errorf("[ERROR] Login with %s user: %v", bootstrapDefaultUser, err)
	}

	log.Printf("[INFO] Connecting with user")
	// Update config token
	return meta.(*Config).UpdateToken(token)

}
