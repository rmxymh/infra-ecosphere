# IPMI Proxy

As a proxy to listen IPMI message, and call the Web API in infra-ecosphere.

This is used in the Virtual Box VM where the provisioning service is located. By this mechanism, IPMI message can hit the IPMI simulation service (infra-ecosphere) and perform the simulated operations.

![image](https://raw.githubusercontent.com/rmxymh/sandbox/master/documents/infra-ecosphere/ipmi-proxy-architecture.png)

In the above image, we have a machine (Host machine) where we runs infra-ecosphere. Besides, we uses Oracle VirtualBox to create a VM and runs Deployment Service (e.g. MAAS) on it. If we want to use infra-ecosphere to simulate BMC so that the deployment service can send IPMI commands to the simulated BMC whose IP address is "127.0.0.0/8".

To solve this issue, IPMI proxy runs on that VM, and it listens on the IP and IPMI port (UDP 623), and convert the IPMI command to REST API call to infra-ecosphere, so that the operation can perform successfully.

## Build

* Build: infra-ecosphere main package is located at ${GOPATH}/src/github.com/rmxymh/infra-ecosphere/infra-ecosphere 

```sh
$ cd ipmi-proxy
$ go install
```


## Usage

Before you run ipmi-proxy, please copy the config file which is used by infra-ecosphere to this VM. For more detailed configuration syntax, you can find from [here](https://github.com/rmxymh/infra-ecosphere/blob/master/README.md#config-file)

After you have prepared for the configuration file in the same path as your executable file, just execute the executable file:

```sh
$ cd $GOBIN
$ ./ipmi-proxy
```

Finally, you can use IPMI command to test it.