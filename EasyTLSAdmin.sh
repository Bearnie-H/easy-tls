#!/usr/bin/env bash

#   Terminal Colour Codes
RED="\e[91m"
GREEN="\e[92m"
YELLOW="\e[93m"
BLUE="\e[96m"
WHITE="\e[97m"
CLEAR="\e[0m"

echo -ne "$WHITE"

if [ -z "$GOPATH" ]; then
    echo -e $RED"ERROR! GOPATH not set!"$WHITE
    stop 1
fi

EasyTLSRepository="github.com/Bearnie-H/easy-tls"
EasyTLSPath="$GOPATH/src/$EasyTLSRepository"

ServerType="server"
ClientType="client"

#   Command-line arguments
Type=
CreatedName=
ProjectFolder=
ModuleFlag=
FrameworkFlag=

#   Which template should be copied over, and to where
TemplateFolder=
DestinationFolder=

function helpMenu() {

    scriptName=$(basename "$0")
    echo -e "
    $BLUE$scriptName - The standard tool for generating a new EasyTLS module or framework from the templates present.$WHITE

    $BLUE$scriptName $YELLOW[-fhmnrpt]$WHITE

    "$YELLOW"Type Flags:$WHITE
        $BLUE-f$WHITE  Framework Flag   -   This script should create a new framework from the template.
                                                This requires the -p and -t flags.
        $BLUE-m$WHITE  Module Flag      -   This script should create a new Module.
                                                This requires the -n, -p and -t flags.
        $BLUE-t$WHITE  Type             -   \"Server\" or \"Client\"

    "$YELLOW"Naming Options:$WHITE
        $BLUE-n$WHITE  CreatedName      -   The name to use for the template.  This defines the folder base 
                                                of the module/framework and root URL for Server modules.
        $BLUE-p$WHITE  Project Path     -   The local path (from $GOPATH/src) to the folder to create the new module in.

    "$YELLOW"Administrative Options:$WHITE
        $BLUE-r$WHITE  Refresh          -   Simply refresh the local copy of the EasyTLS library.

    "$YELLOW"Miscellaneous Options:$WHITE
        $BLUE-h$WHITE  Help Menu        -   Print this help menu and exit,

    New module folders will be located at exactly: 
        $YELLOW\"$GOPATH/src/<Project Path>/<Type>-plugins/<CreatedName>/\"$WHITE

    New framework folders will be located at exactly: 
        $YELLOW\"$GOPATH/src/<Project Path>\"$WHITE
    "

    stop 0
}

function stop() {
    echo -ne "$WHITE"
    exit $1
}

function assertArgSet() {
    local argToCheck="$1"
    local argName="$2"

    if [ -z "$argToCheck" ]; then
        echo -e $RED"ERROR! Required argument \"$argName\" not set!"$WHITE
        helpMenu
        stop 1
    fi
}

function toLower() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | sed 's/[ \_]/-/g'
}

function refreshEasyTLS() {

    echo -e $BLUE"Refreshing EasyTLS library from source..."$WHITE

    go get -u "$EasyTLSRepository"

    if [ $? -eq 0 ]; then
        echo -e $GREEN"Successfully refreshed EasyTLS library.\n"$WHITE
    else
        echo -e $RED"ERROR: Failed to refresh EasyTLS library! \n"$WHITE
    fi
}

function createFrameworkFromTemplate() {

    echo -e $BLUE"Creating new EasyTLS Framework project: 
    $YELLOW$DestinationFolder
    "$WHITE

    #   Refresh the EasyTLS repository
    refreshEasyTLS

    #   Create the new folder
    mkdir -p "$DestinationFolder"

    #   Recursively copy the full contents of the template
    cp -rp "$TemplateFolder/"* "$DestinationFolder/"

    echo -e $GREEN"Successfully created new EasyTLS Framework project:
    $YELLOW$DestinationFolder
    "$WHITE
}

function createModuleFromTemplate() {

    echo -e $BLUE"Creating new EasyTLS module: 
    $YELLOW$DestinationFolder
    "$WHITE

    #   Refresh the EasyTLS repository
    refreshEasyTLS

    #   Create the new folder
    mkdir -p "$DestinationFolder"

    #   Recursively copy the full contents of the template
    cp -rp "$TemplateFolder/"* "$DestinationFolder/"

    echo -e $BLUE"Setting PluginName field of $DestinationFolder/module-definitions.go\n"$WHITE
    #   Define the PluginName field to be the same as the folder-name of the module.
    t=$(cat "$TemplateFolder/module-definitions.go" | sed "s/const PluginName string = \"DEFINE_ME\"/const PluginName string = \"$CreatedName\"/")
    echo "$t" > "$DestinationFolder/module-definitions.go"

    echo -e $GREEN"Successfully created new EasyTLS module:
    $YELLOW$DestinationFolder
    "$WHITE
}

#   Parse the command-line arguments
OPTIND=1
while getopts "fhmn:rp:t:" opt; do
    case "$opt" in
    f)  FrameworkFlag=1
        ;;
    h)  helpMenu
        stop 0
        ;;
    m)  ModuleFlag=1
        ;;
    n)  CreatedName="$OPTARG"
        ;;
    p)  ProjectFolder="$OPTARG"
        ;;
    r)  refreshEasyTLS
        stop 0
        ;;
    t)  Type="$OPTARG"
        ;;
    \?) helpMenu
        stop 1
        ;;
    esac
done

if [ ! -z "$FrameworkFlag" ] && [ ! -z "$ModuleFlag" ]; then
    echo -e $RED"ERROR: -f and -m cannot both be set"$WHITE
    helpMenu
    stop 1
fi

assertArgSet "$Type" "-t"
if ! [[ "$Type" =~ "$ClientType" ]] && ! [[ "$Type" =~ "$ServerType" ]]; then
    echo -e $RED"ERROR: Invalid -t Type.  Must be \"Client\" or \"Server\""$WHITE
    helpMenu
    stop 1
fi

assertArgSet "$ProjectFolder" "-p"

Type=$(toLower "$Type")
ProjectFolder="$GOPATH/src/$ProjectFolder"

if [ ! -z "$FrameworkFlag" ]; then
    TemplateFolder="$EasyTLSPath/examples/example-$Type-framework/"
    DestinationFolder="$ProjectFolder"
    createFrameworkFromTemplate
elif [ ! -z "$ModuleFlag" ]; then
    assertArgSet "$CreatedName" "-n"
    CreatedName=$(toLower "$CreatedName")
    TemplateFolder="$EasyTLSPath/examples/example-$Type-plugin/"
    DestinationFolder="$ProjectFolder/$Type-plugins/$CreatedName"
    createModuleFromTemplate
else
    echo -e $YELLOW"Warning: No framework type selected, nothing has been created."$WHITE
    helpMenu
fi

stop 0