# Build from scratch

The following instructions describe how to build DAOS from source code. The following steps are required:

1. Download Source Code
2. Install Prerequisites
3. Build DAOS
4. Environment Setup

The approximate time to read and execute the steps on this guide is approximately 25 minutes per OS.

## Download Source Code

Download DAOS source code using the following command:

The DAOS repository is hosted on [GitHub](https://github.com/daos-stack/daos).

To checkout the 2.2 development branch, simply run:

```
$ git clone --recurse-submodules https://github.com/daos-stack/daos.git
$ cd daos
```

This command clones the DAOS git repository (path referred as ${daospath}
below) and initializes all the submodules automatically.

## Install Prerequisites

To build DAOS and its dependencies, several software packages must be installed
on the system. This includes scons, libuuid, cmocka, ipmctl, and several other
packages usually available on all the Linux distributions. Moreover, a Go
version of at least 1.10 is required.

Some DAOS tests use MPI. The DAOS build process uses the environment modules
package to detect the presence of MPI. If none is found, the build will skip
building those tests.

Scripts to install all the required packages are provided for each supported
distribution.

### EL (including CentOS)

On CentOS7, please run the following command from the DAOS tree as root or via
sudo.

```bash
$ ./utils/scripts/install-centos7.sh
```

For EL8, the following script must be used instead:

```bash
$ ./utils/scripts/install-el8.sh
```

### openSUSE

For openSUSE, the following command should be executed as root or via sudo:

```bash
$ ./utils/scripts/install-leap15.sh
```

### Unbuntu

As for Ubuntu, please run the following script as the root user or via sudo:

```bash
$ ./utils/scripts/install-ubuntu20.sh
```

## Build DAOS

Once all prerequisites installed and the sources are downloaded,
DAOS can be built via the following command:

```bash
$ scons-3 --config=force --build-deps=yes install
```

By default, DAOS and its dependencies are installed under the `install`
directory.
The installation path can be modified by adding the PREFIX= option to the above
command line (e.g., PREFIX=/usr/local).

!!! note
    Several parameters can be set (e.g., COMPILER=clang or COMPILER=icc) on the
    scons command line. Please see `scons-3 --help` for all the possible options.
    Those options are also saved for future compilations.

## Environment setup

Once built, the environment must be modified to search for binaries and header
files in the installation path. This step is not required if standard locations
(e.g. /bin, /sbin, /usr/lib, ...) are used.

```bash
$ export CPATH=${daospath}/install/include/:$CPATH
$ export PATH=${daospath}/install/bin/:${daospath}/install/sbin:$PATH
```

If using bash, PATH can be set up for you after a build by sourcing the script
utils/sl/setup\_local.sh from the daos root. This script utilizes a file
generated by the build to determine the location of daos and its dependencies.

If required, ${daospath}/install must be replaced with the alternative path
specified through PREFIX.