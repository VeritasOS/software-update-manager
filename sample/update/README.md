# Update plugins' and its dependencies

The document lists the plugins that are bundled into the update RPM using [ĀSUM SDK](../../sdk/README.md).

The update framework workflow details can be found in here: [ĀSUM update framework](../../README.md).

> NOTES

1. The plugins in <font color=red>red color</font> indicates that it was mentioned in dependency, but the plugin itself is missing in the `${PM_LIBRARY}` location.
2. Apart from `.preinstall` plugins, all other plugins can be directly installed into `${PM_LIBRARY}` via component RPM install. So, plugins missing at build time may not be an issue.

## preinstall

![preinstall](./imgs/preinstall.svg)

## install

![install](./imgs/install.svg)

## prereboot

![prereboot](./imgs/prereboot.svg)

## postreboot

![postreboot](./imgs/postreboot.svg)

## rollback-precheck

![rollback-precheck](./imgs/rollback-precheck.svg)

## prerollback

![prerollback](./imgs/prerollback.svg)

## rollback

![rollback](./imgs/rollback.svg)

## commit-precheck

![commit-precheck](./imgs/commit-precheck.svg)

## commit

![commit](./imgs/commit.svg)
