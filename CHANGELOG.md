## 1.1.0 (Unreleased)

ENHANCEMENTS:

* resource/ucloud_eip_association: Update schema version for eip association `ID` [GH-2]
* resource/ucloud_eip_association: Deprecated `resource_type` [GH-2]
* resource/ucloud_lb_attachment: Deprecated `resource_type` [GH-2]
* resource/ucloud_eip: Add `public_ip` attribute [GH-2]
* resource/ucloud_instance: Update `instance_type` about customized [GH-2]
* provider: Add `UserAgent` to external API [GH-2]

BUG FIXES:

* resource/ucloud_disk: Fix default of `name` argument [GH-2]
* resource/ucloud_eip: Fix default of `name` argument [GH-2]
* resource/ucloud_instance: Fix default of `name` argument [GH-2]
* resource/ucloud_lb_listener: Fix default of `name` argument [GH-2]
* resource/ucloud_lb: Fix default of `name` argument [GH-2]
* resource/ucloud_security_group: Fix default of `name` argument [GH-2]
* resource/ucloud_subnet: Fix default of `name` argument [GH-2]
* resource/ucloud_vpc: Fix default of `name` argument [GH-2]

## 1.0.0 (December 19, 2018)

FEATURES:

* **New Resource:** `ucloud_instance`
* **New Resource:** `ucloud_disk`
* **New Resource:** `ucloud_disk_attachment`
* **New Resource:** `ucloud_eip`
* **New Resource:** `ucloud_eip_association`
* **New Resource:** `ucloud_security_group`
* **New Resource:** `ucloud_vpc`
* **New Resource:** `ucloud_subnet`
* **New Resource:** `ucloud_vpc_peering_connection`
* **New Resource:** `ucloud_lb`
* **New Resource:** `ucloud_lb_listener`
* **New Resource:** `ucloud_lb_attachment`
* **New Resource:** `ucloud_lb_rule`
* **New Datasource:** `ucloud_eips`
* **New Datasource:** `ucloud_images`
* **New Datasource:** `ucloud_projects`
* **New Datasource:** `ucloud_zones`
