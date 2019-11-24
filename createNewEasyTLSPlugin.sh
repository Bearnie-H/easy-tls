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

ServerModuleType="server"
ClientModuleType="client"

ModuleType=
ModuleName=

ModuleTemplateFolder=
NewModuleFolder=
ProjectFolder=

function helpMenu() {
    scriptName=$(basename "$0")
    echo -e "
    $GREEN$scriptName - The standard tool for generating a new EasyTLS module from the templates present.$CLEAR

    $scriptName $YELLOW[-hnpt]$CLEAR

        $BLUE-h$CLEAR  Help Menu    -   Print this help menu and exit,
        $BLUE-n$CLEAR  Module Name  -   The name of the module.  This defines the folder base 
                                            of the module and root URL for Server modules.
        $BLUE-r$CLEAR  Refresh      -   Simply refresh the local copy of the EasyTLS library.
        $BLUE-p$CLEAR  Project Path -   The local path (from $GOPATH/src) to the folder to create the new module in.
        $BLUE-t$CLEAR  Module Type  -   \"Server\" or \"Client\"

    The new module folder will be located at exactly: 
        $YELLOW\"$GOPATH/src/<Project Path>/<Module Type>-plugins/<Module Name>/\"$CLEAR
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
    echo "$1" | tr '[:upper:]' '[:lower:]' | sed 's/ \_/-/g'
}

function refreshEasyTLS() {

    echo -e $BLUE"Refreshing EasyTLS library from source..."$CLEAR

    go get -u "github.com/Bearnie-H/easy-tls"

    if [ $? -eq 0 ]; then
        echo -e $GREEN"Successfully refreshed EasyTLS library."$CLEAR
    else
        echo -e $RED"ERROR: Failed to refresh EasyTLS library!"$CLEAR
    fi
} 

function main() {

    #   Refresh the EasyTLS repository
    refreshEasyTLS

    mkdir -p "$NewModuleFolder"

    cp -rp "$ModuleTemplateFolder/"* "$NewModuleFolder/"

    t=$(cat "$ModuleTemplateFolder/module-definitions.go" | sed "s/const PluginName string = \"DEFINE_ME\"/const PluginName string = \"$ModuleName\"/")
    echo "$t" > "$NewModuleFolder/module-definitions.go"
}

#   Parse the command-line arguments
OPTIND=1
while getopts "hn:p:rt:" opt; do
    case "$opt" in
    h)  helpMenu
        ;;
    n)  ModuleName=$OPTARG
        ;;
    p)  ProjectFolder=$OPTARG
        ;;
    r)  refreshEasyTLS
        exit 0
        ;;
    t)  ModuleType=$OPTARG
        ;;
    \?) helpMenu
        ;;
    esac
done

assertArgSet "$ModuleName"    "-n"
assertArgSet "$ProjectFolder" "-p"
assertArgSet "$ModuleType"    "-t"

ModuleName=$(toLower "$ModuleName")
ProjectFolder=$(toLower "$ProjectFolder")
ModuleType=$(toLower "$ModuleType")

ProjectFolder="$GOPATH/src/$ProjectFolder/"
NewModuleFolder="$ProjectFolder/$ModuleType-plugins/$ModuleName"
ModuleTemplateFolder="$EasyTLSPath/example-$ModuleType-plugin/"

main