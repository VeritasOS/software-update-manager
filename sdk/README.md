# SUM SDK

## Update RPM Generation

An update RPM should be SUM format compliant in order for one to successfully install the update. In order to generate such an RPM, one should use the [SUM SDK](../sdk/Makefile) along with providing appropriate [rpm information](./rpm-info.json).

### Prerequisites

To use this framework, you need the following:

1. SUM SDK
2. JSON file containing RPM information. Ex: [rpm-info.json](./rpm-info.json).
3. Library of plugin folders containing plugins.
4. Optionally ISO can be specified through `ISO_PATH` to include ISO contents into the RPM.
   NOTE: The ISO contents would be extracted into `$(PLUGINS_LIBRARY)/iso/contents` folder, and appropriate `.install` plugin should be included to install the contents of this ISO to do online or offline upgrade. One can place the  required `.install` plugins inside `$(PLUGINS_LIBRARY)/iso` or any other plugin folder, and access the contents of ISO using `$(PLUGINS_LIBRARY)/iso/contents` path.

### Usage

Download and extract the SUM SDK, and call the `make` target as below by specifying appropriate parameters as shown below:

```bash
$(MAKE) -C $(SUM_SDK_PATH) generate \
    ISO_PATH=$${Path_to_ISO} \
    PLUGINS_LIBRARY=$${PluginsLibraryPath} \
    RPM_NAME=$${RPM_Name} \
    RPM_VERSION=$${Version_of_RPM} \
    RPM_REVISION=$${Revision_of_RPM} \
    RPM_URL=$${Update_RPM} \
    RPM_SUMMARY=$${Summary_of_the_RPM} \
    RPM_INFO_FILE=$${Update_RPM_Info_File} \
    SHIP_DIR=$${RPM_Destination_Dir};
```
