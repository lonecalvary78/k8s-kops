# Flannel

&#9888; The Flannel CNI is not supported for Kubernetes 1.28 or later.

## Installing

To install [flannel](https://github.com/coreos/flannel) - use `--networking flannel-vxlan` (recommended) or `--networking flannel-udp` (legacy).  `--networking flannel` now selects `flannel-vxlan`.

```sh
export ZONES=mylistofzone
kops create cluster \
  --zones $ZONES \
  --networking flannel \
  --yes \
  --name myclustername.mydns.io
```

## Configuring

### Configuring Flannel iptables resync period

As of kOps 1.12.0, Flannel iptables resync option is configurable via editing a cluster and adding
`iptablesResyncSeconds` option to spec:

```yaml
  networking:
    flannel:
      iptablesResyncSeconds: 360
```