# Service Fabric Emulator

*Working in process*

Launch stateful/stateless service fabric applications anywhere. No service fabric cluster required.

## Motivation

Service Fabric provides many cloud features such as leader election (Primary role) and cloud storage (Reliable Collection). However, in order to use those amazing futures, a very heavy SDK and runtime have to be installed. It is extremely difficult for Service Fabric users to migrate to other clouds or even leverage from other technology not compatible with Service Fabric.

A better approach would be [Dapr](https://dapr.io/). But no one wants to pay the price of migrating from old school service fabric apps.

Service Fabric Emulator keeps your apps binary the same, no code change required. (good news to anyone who lost the source code)

The _Emulator_ is like the emulator of NES on the PC. All be loved games should run smoothly with the Xbox controllers.

## Architecture 

Components remain the same:
 
 * `USER APP`: the application package built for service fabric cluster
 * `SF SDK`: the lib imported by the app. typically from nuget if .net app or maven if java app
 * `FabricRuntime.dll`: the runtime installed by service fabric. this is a native lib which exposes APIs and interop with `SF SDK`. An alternative way to emulate service fabric is to replace this layer with a customized version runtime. however, the open source code of service fabric does not work well on Windows and a bit of outdate. This layer may subject to be `emulated` in the future.

Components to be emulated:

 * `Fabric.exe` -> `Emulator`: This layer listens on a tcp port and accepts connections from `SF SDK` that would send some commands via the connections. for example, save data to the could. The `Emulator` implements the commands from `SF SDK` and acts as the local agent to the cloud.
 * `Fabric Cluster` -> `Any Cloud`: The `Emulator` is designed to be a API layer to fit any cloud. This can be a kuberentes leader election or a storage engine backed by MySQL to emulate the reliable collection in the app.


```
                                                      +
                                                      |
              Traditional SF APP                      |                        SF Emulator
                                                      |
                                                      |
                                                      |
                                                      |
         +---------------------------------+          |              +---------------------------------+
         |                                 |          |              |                                 |
         |     USER APP                    |          |              |     USER APP                    |
         |                                 |          |              |                                 |
         |                                 |          |              |                                 |
         |                                 |          |              |                                 |
         |         +-------------------+   |          |              |         +-------------------+   |
         |         |                   |   |          |              |         |                   |   |
         |         |      SF SDK       |   |          |              |         |      SF SDK       |   |
         |         |                   |   |          |              |         |                   |   |
         |         +----------+--------+   |          |              |         +----------+--------+   |
         |                    |            |          |              |                    |            |
         +---------------------------------+          |              +---------------------------------+
                              |                       |                                   |
                              |                       |                                   |
                              |                       |                                   |
             +----------------v----------+            |                  +----------------v----------+
             |                           |            |                  |                           |
             |      FabricRuntime.dll    |            |                  |      FabricRuntime.dll    |
             |                           |            |                  |                           |
             +------------+--------------+            |                  +------------+--------------+
                          |                           |                               |
+---------------------------------------------------------------------------------------------------------------+
                          |                           |                               |
             +------------v--------------+            |                  +------------v--------------+
             |                           |            |                  |                           |
             |                           |            |                  |                           |
             |                           |            |                  |                           |
             |         Fabric.exe        |            |                  |      Fabric Emulator      |
             |                           |            |                  |                           |
             |                           |            |                  |                           |
             |                           |            |                  |                           |
             +------------^--------------+            |                  +------------^--------------+
                          |                           |                               |
                          |                           |                               |
                          |                           |                               |
          +---------------v------------------+        |               +---------------v------------------+
          |                                  |        |               |                                  |
          |                                  |        |               |                                  |
          |                                  |        |               |                                  |
          |                                  |        |               |        Any cloud provider        |
          |         Fabric Cluster           |        |               |                                  |
          |                                  |        |               |         e.g. Kubernetes          |
          |                                  |        |               |                                  |
          |                                  |        |               |                                  |
          +----------------------------------+        |               +----------------------------------+
                                                      |
                                                      |
                                                      |
                                                      +

```

## Supported features

 * Change Role, see [changerole](examples/changerole/)

## Planned features

 * Reliable Collections

