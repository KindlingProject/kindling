/*

Copyright (c) 2013-2018 Draios Inc. dba Sysdig.

This file is dual licensed under either the MIT or GPL 2. See MIT.txt
or GPL2.txt for full copies of the license.

*/
#pragma once

#define PROBE_VERSION "0.1.1dev"

#define PROBE_NAME "hcmine-probe"

#define PROBE_DEVICE_NAME "hcmine"

#ifndef KBUILD_MODNAME
#define KBUILD_MODNAME PROBE_NAME
#endif
