# Presto Admin Pool for Reeve

The Presto Admin pool contain a set of scripts to install, uninstall, configure, get status and manage (start, stop, restart) Presto Server on a Reeve cluster.

Once the Reeve cluster exists (with at least a node), when a new node is created and joint to the Reeve cluster, **if the node is also part of the Presto Cluster**, the Presto Server will be installed, configured and started. To be part of the Presto Cluster the Reeve node should have the tag **presto_role** with the value **coordinator** or **worker**. To There is no need to execute another command to have a new Presto node ready, unless there is an error :).

Other information that should receive every node, except the Reeve Master, is the hostname or IP of the Reeve Master. It is received in the **REEVE_MASTER** environment variable. **REEVE_SERVICE_AS** is other environment variable with the kind of service to configure Reeve, the values are: sysv, systemd or script. By default (if the variable is not provided) is sysv.

It is important to assign this tag when the new node is set. This is done by having environment variables starting with **REEVE_TAG_**. For example, if you are using Docker Compose include the `environment` section with the tags to assing:

    environment:
      REEVE_SERVICE_AS: sysv
      REEVE_MASTER: reeve_master
      REEVE_TAG_MYSQL_ROLE: master
      REEVE_TAG_PRESTO_ADAPTER: mysql

**NOTE**: The Reeve bootrap script works with these environment variables now, but this will be a function of Reeve, taking the environment varaibles, with the parameter `--tag` or a config file.

## Directories and Files

Besides the required directories in the Presto pool (member, query and user) there are other directories used by the event handlers (not Reeve) and the user or Presto Server admin.

* **presto/bin**: Contain scripts to execute actions instead of use Serf.

  * **update_coordinator.sh**: To update Presto Server to a newer version it can be done with the `presto/update` event but it is only for the Workers. To update the Coordinator and the Workers execute `bin/update_coordinator.sh` after modify the install parameters in `config/*.conf`. This script will update Presto Server at the Coordinator then execute the update event to do the same in the Workers.

  * **verifier.sh**: This script will install the verifier JAR file and create a config properties file to verify the MySQL Adapter.

* **presto/config**: Contain default configuration parameters for the Coordinator, Workers or both. These would be loaded when a event handler is executed or when it is not possible to get the parameters from the Coordinator (default parameters)

  * **global.conf**: Default configuration parameters for both, the Coordinator and Workers. For example, install and configuration parameters.
  * **coordinator.conf** and **worker.conf**: Default configuration parameters for the Coordinator or the Worker.

* **presto/src**: Source code of the event handlers. Some event handlers are developed in a scripting language such as Bash or Python and the code is located inside the event type directory. But for those developed in Go require the source code to be located in `presto/src`, once they are compiled the binary is placed in the event type directory.

* **presto/{user,query,member}**: The event type directories. They have the event handlers. Read below for more information about each one.

* **Makefile**: All the `make` actions to automate the use and developent of this pool.

## Using Presto Admin

You can also use make from the Docker directory to execute Presto Admin commands.

To show the status of every Presto node use `presto-status`. This will show the version of Java, Presto Server (if they are installed), if Presto Server is running and more information.

    make presto-status

Use `presto-dashboard` to open the Presto Server dashboard with information about the workers, queries, etc...

    make presto-dashboard

More commands can be executed inside the node using Serf or at your computer using `make event` and `make query`.

### Install and Configure when a Worker join the Presto Cluster

Once a node is created it will join to the Reeve cluster and it will install, configure and start Presto Server. This is done by the `member/join` event, it is execute once a Reeve node join or re-join a Reeve cluster.

If the node is in the Presto Cluster and it is the new guy in the cluster (if it is already there, do nothing) it will:
1. Load the install parameters from the Coordinator (if it is a Worker) or locally (if it is the Coordinator) using the query `presto/installdata`
2. Download Java from Oracle to Install or Upgrade it if required.
3. Download from Teradata Presto Server to Install or Upgrade it if required.
4. Load the default config parameters from the Coordinator (if it is a Worker) or locally (if it is the Coordinator)
5. Create the `/etc/presto/config.properties` file with the appropiate values if the node is a Coordinator or Worker.
6. Startup Presto Server

If the node is in the Presto Cluster and the new node in the cluster is an adapter, it will:
1. A


### Update

Use the event `presto/update` when there is a new version of Presto Server or Java to install/update in every Presto Cluster Worker. First modify the `presto/config/global.conf` or `presto/config/*.conf` with the new installation parameters (i.e. version, filename or url) and send the event.

The event will apply the update to every worker, so to update the Coordinator and the Workers use the script  `presto/bin/update_coordinator.sh`:

    # Update all the Workers:
    serf event presto/update

    # Update all the Presto nodes:
    ./presto/bin/update_coordinator.sh

This event will be executed only if the Reeve node is in the Presto Cluster and if it is a Presto Worker (presto_role == worker), then it will:
1. Identify the Coordinator and send it a `presto/installdata` query to load the install parameters.
2. Download Java from Oracle to Install or Upgrade it if required.
3. Download from Teradata Presto Server to Install or Upgrade it if required.
4. Get the configuration data from the Coordinator (using `presto/configdata`) and apply the change if there is something new.
5. Restart Presto Server if something was installed or updated.

The script `presto/bin/update_coordinator.sh` will do basically the same but instead of request the install/config data with `presto/installdata`/`presto/configdata` (steps #1 and #4) it will load it from the configuration files `presto/config/global.conf` or `presto/config/coordinator.conf`. Once Java and Presto Server are installed or updated, it will send the event `presto/update` to command the Workers to update (step #3.1).

## Installation Options

With the `member/join` and `presto/update` events we can install Presto Server. The `member/join` will install it when a new node join into the Presto cluster, it is a self-install process, no human action is required.

The `presto/update` is an install on demand process. The Coordinator will command to every Worker to install or update Presto Server (it's not force, the Worker will do it if required). It will also change the configuration and restart if it is required.

## Configure or Deploy a Configuration Change

To deploy new configuration changes use the event `presto/config`. Use it to re-configure the node when the initial configuration fail or to apply/deploy a new configuration change. All the configuration values are located in `presto/config/global.conf` or `presto/config/*.conf`. If there was a change in the configuration file, Presto Server will be restarted.

    serf event presto/config

If the node is in the Presto Cluster, the event handler will:
1. Load the configuration data from the Coordinator using `presto/configdata` (if it is a Worker) or locally from the config files `presto/config/global.conf` and `presto/config/coordinator.conf` (if it is the coordinator)
2. Create dynamic parameters from the loaded.
3. Create a temporal configuration file. If the file differs from the original, then will replace it and restart Presto Server.

## Status

The `presto/status` query will provide information about the Presto Server. At this time it is providing:
* Java Oracle version or if it is not installed.
* Presto Server version or if it is not installed.
* Presto Server status (Running or Not Running) and Process ID.
* Presto Role (Coordinator or Worker)

More information will be included.

## Presto Server Management (Start, Stop & Restart)

Presto Server can be managed remotely using the following events:

* Start: `presto/start`
* Stop: `presto/stop`
* Restart: `presto/restart`

They will use the Presto service `/etc/init.d/presto` and return (visible only at Serf logs) the output which include the Process ID.

## Uninstall

The `presto/uninstall` query will uninstall Presto Server and Java Oracle RPM's. It is a query and not an event because events cannot be directed to an specific node, queries can.

In order to uninstall Presto Server in a specific node or limited group of nodes, include the parameter `--node` with the node name. To know the node names, you can use the `serf members` command.

Example:

    # Identify the nodes to uninstall
    serf members

    # Uninstall Presto at Worker #2
    serf -node reeve-WORKER02.docker.co query presto/uninstall

    # Uninstall Presto at Worker #1 & #2
    serf -node reeve-WORKER01docker.co -node reeve-WORKER02.docker.co query presto/uninstall

The uninstall functionality should not be included in the `member/leave` or `member/failed` event handlers because a node may be out of the cluster by an error and not intentionally. To intentionally remove a node from the cluster, you first have to uninstall Presto Server (if desired), then remove it.

## TODO

- [X] Join to install, configure and start Presto when a node join the cluster. DONE
- [X] Update to update/install Presto commanded by the Presto Coordinator. DONE
- [X] Config to configure or update the configuration of the Presto node commanded by the Presto Coordinator. DONE
- [X] InstallData and ConfigData to provide the latest versions or parameters. DONE
- [X] Status to provide information about Presto Server. Done: But need to include more information about the system.
- [X] Start, Stop and Restart for Presto Server. DONE
- [X] Uninstall.
- [ ] Add/Remove a Connector
- [ ] Collect logs
- [ ] Collect query_info
- [ ] Include system info in Status
- [ ] Show confguration (in Status?)
- [ ] Migrate all or most of the event handlers to Pyhton or Go
