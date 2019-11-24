#!/usr/bin/env bash

#   Terminal Colour Codes
RED="\e[91m"
GREEN="\e[92m"
YELLOW="\e[93m"
BLUE="\e[96m"
CLEAR="\e[0m"

if [ -z "$GOPATH" ]; then
    echo -e "\e[91mERROR! GOPATH not set!\e[0m"
    exit 1
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
    $GREEN$scriptName - The standard tool for generating a new EasyTLS module or framework from the templates present.$CLEAR

    $scriptName $YELLOW[-fhmnrpt]$CLEAR

        $BLUE-f$CLEAR  Framework Flag   -   This script should create a new framework from the template.
                                                This requires the -p and -t flags.
        $BLUE-h$CLEAR  Help Menu        -   Print this help menu and exit,
        $BLUE-m$CLEAR  Module Flag      -   This script should create a new Module.
                                                This requires the -n, -p and -t flags.
        $BLUE-n$CLEAR  CreatedName      -   The name to use for the template.  This defines the folder base 
                                                of the module/framework and root URL for Server modules.
        $BLUE-r$CLEAR  Refresh          -   Simply refresh the local copy of the EasyTLS library.
        $BLUE-p$CLEAR  Project Path     -   The local path (from $GOPATH/src) to the folder to create the new module in.
        $BLUE-t$CLEAR  Type             -   \"Server\" or \"Client\"

    New module folders will be located at exactly: 
        $YELLOW\"$GOPATH/src/<Project Path>/<Type>-plugins/<CreatedName>/\"$CLEAR

    New framework folders will be located at exactly: 
        $YELLOW\"$GOPATH/src/<Project Path>\"$CLEAR
    "

    exit 0
}

function assertArgSet() {
    local argToCheck="$1"
    local argName="$2"

    if [ -z "$argToCheck" ]; then
        echo -e $RED"ERROR! Required argument \"$argName\" not set!"$CLEAR
        helpMenu
    fi
}

function toLower() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | sed 's/[ \_]/-/g'
}

function refreshEasyTLS() {

    echo -e $BLUE"Refreshing EasyTLS library from source..."$CLEAR

    go get -u "github.com/Bearnie-H/easy-tls"

    if [ $? -eq 0 ]; then
        echo -e $GREEN"Successfully refreshed EasyTLS library.\n"$CLEAR
    else
        echo -e $RED"ERROR: Failed to refresh EasyTLS library! \n"$CLEAR
    fi
}

function createFrameworkFromTemplate() {

    echo -e $BLUE"Creating new EasyTLS Framework project: 
    $YELLOW$DestinationFolder
    "$CLEAR

    #   Refresh the EasyTLS repository
    refreshEasyTLS

    #   Create the new folder
    mkdir -p "$DestinationFolder"

    #   Recursively copy the full contents of the template
    cp -rp "$TemplateFolder/"* "$DestinationFolder/"

    echo -e $GREEN"Successfully created new EasyTLS Framework project:
    $YELLOW$DestinationFolder
    "$CLEAR
}

function createModuleFromTemplate() {

    echo -e $BLUE"Creating new EasyTLS module: 
    $YELLOW$DestinationFolder
    "$CLEAR

    #   Refresh the EasyTLS repository
    refreshEasyTLS

    #   Create the new folder
    mkdir -p "$DestinationFolder"

    #   Recursively copy the full contents of the template
    cp -rp "$TemplateFolder/"* "$DestinationFolder/"

    echo -e $BLUE"Setting PluginName field of $DestinationFolder/module-definitions.go\n"$CLEAR
    #   Define the PluginName field to be the same as the folder-name of the module.
    t=$(cat "$TemplateFolder/module-definitions.go" | sed "s/const PluginName string = \"DEFINE_ME\"/const PluginName string = \"$CreatedName\"/")
    echo "$t" > "$DestinationFolder/module-definitions.go"

    echo -e $GREEN"Successfully created new EasyTLS module:
    $YELLOW$DestinationFolder
    "$CLEAR
}

#   Parse the command-line arguments
OPTIND=1
while getopts "fhmn:rp:t:" opt; do
    case "$opt" in
    f)  FrameworkFlag=1
        ;;
    h)  helpMenu
        ;;
    m)  ModuleFlag=1
        ;;
    n)  CreatedName="$OPTARG"
        ;;
    p)  ProjectFolder="$OPTARG"
        ;;
    r)  refreshEasyTLS
        exit 0
        ;;
    t)  Type="$OPTARG"
        ;;
    \?) helpMenu
        ;;
    esac
done

if [ ! -z "$FrameworkFlag" ] && [ ! -z "$ModuleFlag" ]; then
    echo -e $RED"ERROR: -f and -m cannot both be set"$CLEAR
    helpMenu
fi

assertArgSet "$Type" "-t"
if ! [[ "$Type" =~ "$ClientType" ]] && ! [[ "$Type" =~ "$ServerType" ]]; then
    echo -e $RED"ERROR: Invalid -t Type.  Must be \"Client\" or \"Server\""$CLEAR
    helpMenu
fi

assertArgSet "$ProjectFolder" "-p"

Type=$(toLower "$Type")
ProjectFolder=$(toLower "$GOPATH/src/$ProjectFolder")

if [ ! -z "$FrameworkFlag" ]; then
    TemplateFolder="$EasyTLSPath/examples/example-$Type-framework/"
    DestinationFolder="$ProjectFolder"
    createFrameworkFromTemplate
    exit 0
elif [ ! -z "$ModuleFlag" ]; then
    assertArgSet "$CreatedName" "-n"
    CreatedName=$(toLower "$CreatedName")
    TemplateFolder="$EasyTLSPath/examples/example-$Type-plugin/"
    DestinationFolder="$ProjectFolder/$Type-plugins/$CreatedName"
    createModuleFromTemplate
    exit 0
fi  

echo -e $YELLOW"Warning: No framework type selected, nothing has been created."$CLEAR
helpMenu

