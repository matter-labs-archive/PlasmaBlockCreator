# Description

This repository contains a block production part of the Plasma. The list of responsibilities of this software is the following:
- Accept new transactions if signature and UTXOs being spent are valid
- Resolve double-spends (simultaneous spam attempt to spend twice)
- Keep linearizable order
- Assemble blocks
- Write new UTXOs from assembled blocks
- Process deposit events
- Process exit events and provide information for challenges if necessary (for cases when user tries to exit an already spent UTXO)
- Process requests for "deposit withdraw" and provide info for challenge in case if the deposit was already processed

Description of the public Web API that PlasmaBlockCreator provides is described [here](https://matterinc.github.io/PlasmaBlockCreator/)

At the moment it's a large monolith part for ease of integrational testing and will be later slit for modules with small set of well-defined functionality.

Code here is updated and sometimes uploaded to the Docker store. For deployment scripts please refer to this [repo](https://github.com/matterinc/DeploymentTools).

### Authors

- Alex Vlasov, [@shamatar](https://github.com/shamatar)
