# Update plugins' and its dependencies

The document lists the plugins that are bundled into the update RPM using [SUM SDK](https://github.com/VeritasOS/software-update-manager/blob/v1/sdk/README.md).

The update framework workflow details can be found in here: [SUM update framework](https://github.com/VeritasOS/software-update-manager/blob/v1/README.md).

> NOTES

1. The plugins in <font color=red>red color</font> indicates that it was mentioned in dependency, but the plugin itself is missing in the `${PM_LIBRARY}` location.
2. Apart from `.preinstall` plugins, all other plugins can be directly installed into `${PM_LIBRARY}` via component RPM install. So, plugins missing at build time may not be an issue.
