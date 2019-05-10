package local

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceLocalFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocalFileCreateUpdate,
		Read:   resourceLocalFileRead,
		Delete: resourceLocalFileDelete,
		Update: resourceLocalFileCreateUpdate,

		Schema: map[string]*schema.Schema{
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"sensitive_content"},
			},
			"sensitive_content": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"content"},
			},
			"filename": {
				Type:        schema.TypeString,
				Description: "Path to the output file",
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceLocalFileRead(d *schema.ResourceData, _ interface{}) error {
	// If the output file doesn't exist, mark the resource for creation.
	outputPath := d.Get("filename").(string)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		d.SetId("")
		return nil
	}

	// Verify that the content of the destination file matches the content we
	// expect. Otherwise, the file might have been modified externally and we
	// must reconcile.
	outputContent, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return err
	}

	if string(outputContent) != resourceLocalFileContent(d) {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceLocalFileContent(d *schema.ResourceData) string {
	content := d.Get("content")
	sensitiveContent, sensitiveSpecified := d.GetOk("sensitive_content")
	useContent := content.(string)
	if sensitiveSpecified {
		useContent = sensitiveContent.(string)
	}

	return useContent
}

func resourceLocalFileCreateUpdate(d *schema.ResourceData, _ interface{}) error {
	content := resourceLocalFileContent(d)
	destination := d.Get("filename").(string)

	destinationDir := path.Dir(destination)
	if _, err := os.Stat(destinationDir); err != nil {
		if err := os.MkdirAll(destinationDir, 0777); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(destination, []byte(content), 0777); err != nil {
		return err
	}

	d.SetId("-")

	return nil
}

func resourceLocalFileDelete(d *schema.ResourceData, _ interface{}) error {
	os.Remove(d.Get("filename").(string))
	return nil
}
