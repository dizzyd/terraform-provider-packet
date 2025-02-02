package packet

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/packethost/packngo"
)

func TestAccPacketSSHKey_Basic(t *testing.T) {
	var key packngo.SSHKey
	rInt := acctest.RandInt()
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPacketSSHKeyConfig_basic(rInt, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketSSHKeyExists("packet_ssh_key.foobar", &key),
					resource.TestCheckResourceAttr(
						"packet_ssh_key.foobar", "name", fmt.Sprintf("foobar-%d", rInt)),
					resource.TestCheckResourceAttr(
						"packet_ssh_key.foobar", "public_key", publicKeyMaterial),
					resource.TestCheckResourceAttrSet(
						"packet_ssh_key.foobar", "owner_id"),
				),
			},
		},
	})
}

func TestAccPacketSSHKey_ProjectBasic(t *testing.T) {
	rInt := acctest.RandInt()
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPacketSSHKeyConfig_projectBasic(rInt, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"packet_project.test", "id",
						"packet_project_ssh_key.foobar", "project_id",
					),
				),
			},
		},
	})
}

func TestAccPacketSSHKey_Update(t *testing.T) {
	var key packngo.SSHKey
	rInt := acctest.RandInt()
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPacketSSHKeyConfig_basic(rInt, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketSSHKeyExists("packet_ssh_key.foobar", &key),
					resource.TestCheckResourceAttr(
						"packet_ssh_key.foobar", "name", fmt.Sprintf("foobar-%d", rInt)),
					resource.TestCheckResourceAttr(
						"packet_ssh_key.foobar", "public_key", publicKeyMaterial),
				),
			},
			{
				Config: testAccCheckPacketSSHKeyConfig_basic(rInt+1, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketSSHKeyExists("packet_ssh_key.foobar", &key),
					resource.TestCheckResourceAttr(
						"packet_ssh_key.foobar", "name", fmt.Sprintf("foobar-%d", rInt+1)),
					resource.TestCheckResourceAttr(
						"packet_ssh_key.foobar", "public_key", publicKeyMaterial),
				),
			},
		},
	})
}

func TestAccPacketSSHKey_projectImportBasic(t *testing.T) {
	sshKey, _, err := acctest.RandSSHKeyPair("")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPacketSSHKeyConfig_projectBasic(acctest.RandInt(), sshKey),
			},
			{
				ResourceName:      "packet_project_ssh_key.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPacketSSHKey_importBasic(t *testing.T) {
	sshKey, _, err := acctest.RandSSHKeyPair("")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPacketSSHKeyConfig_basic(acctest.RandInt(), sshKey),
			},
			{
				ResourceName:      "packet_ssh_key.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPacketSSHKeyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*packngo.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "packet_ssh_key" {
			continue
		}
		if _, _, err := client.SSHKeys.Get(rs.Primary.ID, nil); err == nil {
			return fmt.Errorf("SSH key still exists")
		}
	}

	return nil
}

func testAccCheckPacketSSHKeyExists(n string, key *packngo.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*packngo.Client)

		foundKey, _, err := client.SSHKeys.Get(rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if foundKey.ID != rs.Primary.ID {
			return fmt.Errorf("SSh Key not found: %v - %v", rs.Primary.ID, foundKey)
		}

		*key = *foundKey

		fmt.Printf("key: %v", key)
		return nil
	}
}

func testAccCheckPacketSSHKeyConfig_basic(rInt int, publicSshKey string) string {
	return fmt.Sprintf(`
resource "packet_ssh_key" "foobar" {
    name = "foobar-%d"
    public_key = "%s"
}`, rInt, publicSshKey)
}

func testAccCheckPacketSSHKeyConfig_projectBasic(rInt int, publicSshKey string) string {
	return fmt.Sprintf(`

resource "packet_project" "test" {
    name = "test-%d"
}

resource "packet_project_ssh_key" "foobar" {
    name = "foobar-%d"
    public_key = "%s"
	project_id = packet_project.test.id
}`, rInt, rInt, publicSshKey)
}
