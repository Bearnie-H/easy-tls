#!/usr/bin/env bash

#   Author: 
#   Date:   

#   Purpose:
#
#       This script is intended to be the generic build-script for an EasyTLS plugin.

#   Terminal Colour Codes
RED="\e[91m"
GREEN="\e[92m"
YELLOW="\e[93m"
BLUE="\e[96m"
WHITE="\e[97m"
CLEAR="\e[0m"

#   Assert that the "default" background colour is white.  This should be used as the background colour within the script, so as to provide a consistent visual language with colour.
echo -ne "$WHITE"

if [ -z "$GOPATH" ]; then
    echo -e $RED"ERROR: GOPATH environment variable not set!"$WHITE
    stop 1
fi

#   Command-line / Global variables
artefactDirectory="$GOPATH/src/easytls-compiled-plugins"

BuildModeProduction="prod"
BuildModeDebug="debug"

ModuleTypeClient="client"
ModuleTypeServer="server"

ForceMode=
ModuleType=
BuildMode=
BuildHash=
RaceDetector=""

#   Function to display a help/usage menu to the user in a standardized format.
function helpMenu() {

    #   Get the name of the executable
    scriptName=$(basename "$0")

    #   Print the current help menu/usage information to the user
    echo -e "
    $BLUE$scriptName   -   A bash tool to build an EasyTLS plugin in a consistent and generic manner.$WHITE

    $GREEN$scriptName $YELLOW[-fhmrx]$WHITE

    "$YELLOW"Build Options:$WHITE
        $BLUE-f$WHITE  -   Force Mode. Build the plugin, ignoring the presence or absence
                                of the standard \"RELEASE\" flag in the plugin directory.
        $BLUE-m$WHITE  -   Build Mode. Should the plugin be build in \"prod\" or \"debug\" mode.
        $BLUE-r$WHITE  -   Race Detector. Include the Golang race detector in the compilation.
        $BLUE-x$WHITE  -   Build Hash. A hash code to allow multiple concurrent builds without stamping outputs.

    "$YELLOW"Miscellaneous Options:$WHITE
        $BLUE-h$WHITE  -   Display this help menu and exit.
    "
}

function stop() {
    echo -ne "$CLEAR"
    exit $1
}

#   Helper function to allow asserting that required arguments are set.
function argSet() {
    local argToCheck="$1"
    local argName="$2"

    if [ -z "$argToCheck" ]; then
        echo -e $RED"ERROR! Required argument \"$argName\" not set!"$WHITE
        helpMenu
        stop 1
    fi
}

#   Helper function to allow checking for the existence of files on disk.
function fileExists() {
    local FilenameToCheck="$1"

    if [ ! -f "$FilenameToCheck" ]; then
        echo -e $YELLOW"Warning: File \"$FilenameToCheck\" does not exist."$WHITE
        return 1
    fi

    return 0
}

#   Helper function to allow checking for the existence of directories on disk.
function directoryExists() {
    local DirectoryToCheck="$1"

    if [ ! -d "$DirectoryToCheck" ]; then
        echo -e $YELLOW"Warning: Directory \"$DirectoryToCheck\" does not exist."$WHITE
        return 1
    fi

    return 0
}

function incrementBuildCount() {
    if [ ! -f version.go ]; then
        echo -e $RED"ERROR! Failed to find required \"version.go\" file in $(pwd)!"$CLEAR
        stop 1
    fi
    local lastBuildCount=$(cat version.go | grep 'Build' | sed 's/[ \t,]*//g' | awk -F: '{print $2}');
    lastBuildCount=$(( $lastBuildCount + 1 ))
    sed -i "s/Build\:[ \t]*[0-9]*/Build:$lastBuildCount/g" version.go 
    gofmt -s -w version.go
}

function ShouldRun() {

    read -p "Build $ModuleType Module \"$(basename $(pwd)\")? [y/N]: " response

    if [[ "$response" =~ [yY] ]]; then
        echo "Yes"
    fi
}

function setCycleTimes() {

    #   Production mode
    if [[ "$BuildMode" =~ "$BuildModeProduction" ]]; then
        sed -i 's/var DefaultPluginCycleTime time.Duration =.*$/var DefaultPluginCycleTime time.Duration = time.Minute \* 5/' standard-definitions.go

    #   Debug mode
    elif [[ "$BuildMode" =~ "$BuildModeDebug" ]]; then
        sed -i 's/var DefaultPluginCycleTime time.Duration =.*$/var DefaultPluginCycleTime time.Duration = time.Second \* 15/' standard-definitions.go
    fi

    gofmt -s -w standard-definitions.go
}

function main() {

    artefactDirectory="$artefactDirectory-$BuildHash"

    echo -e $BLUE"\nExecuting $0"$CLEAR

    buildDir=$(cd "$(dirname "$0")"; pwd)

    cd "$buildDir"

    if [ ! -z $(echo "$buildDir" | grep -i "\/client-plugins") ]; then
        ModuleType="Client"
    elif [ ! -z $(echo "$buildDir" | grep -i "\/server-plugins") ]; then
        ModuleType="Server"
    fi

    if [ -z "$ForceMode" ]; then
        if [ ! -e "RELEASE" ]; then
            echo -e $YELLOW"WARNING: Module $(basename $buildDir) not marked for release. Skipping..."$CLEAR
            stop 0
        else
            shouldRun=$(ShouldRun)
            if [ -z "$shouldRun" ]; then
                echo -e $BLUE"Skipping build..."$CLEAR
                stop 0
            fi
            echo -e $GREEN"Building Module $(basename $buildDir)..."$CLEAR
        fi
    fi

    echo -e $BLUE"Updating Module Build Number..."$CLEAR
    incrementBuildCount
    echo -e $GREEN"Finished Updating Module Build Number."$CLEAR

    if [[ "$ModuleType" =~ "Client" ]]; then
        echo -e $BLUE"Setting plugin cycle times for build mode: $BuildMode"$CLEAR
        setCycleTimes
        echo -e $GREEN"Successfully set plugin cycle times for build mode: $BuildMode"$CLEAR
    fi

    pluginName=$(basename "$buildDir")
    if [ ! -z $(echo "$buildDir" | grep 'server') ]; then
        pluginName="server-$pluginName"
    else
        pluginName="client-$pluginName"
    fi

    if [ ! -z "$RaceDetector" ]; then
        echo -e $BLUE"Building with Race Detection: [ "$WHITE"enabled"$BLUE" ]"$WHITE
    else
        echo -e $BLUE"Building with Race Detection: [ "$WHITE"disabled"$BLUE" ]"$WHITE
    fi

    go clean -modcache
    go clean -cache
    go clean

    echo -e $BLUE"Compiling plugin..."$CLEAR
    go build -buildmode=plugin -o "$artefactDirectory/$pluginName.so" -gcflags="all=-N -l" "$RaceDetector"
    if [ $? -eq 0 ]; then
        echo -e $GREEN"Built Module $(basename $buildDir). Artefact located at \"$artefactDirectory/$pluginName.so\" "$CLEAR
        cd - > /dev/null
    else
        echo -e $RED"ERROR! Failed to Build Module $(basename $buildDir)."$CLEAR
        cd - > /dev/null
        stop 1
    fi

    echo -e $BLUE"Finished Executing $0"$CLEAR
}


#   Parse the command line arguments.  Add the flag name to the list (in alphabetical order), and add a ":" after if it requires an argument present.
#   The value of the argument will be located in the "$OPTARG" variable
while getopts "fhm:rx:" opt; do
    case "$opt" in
    f)  ForceMode=1
        ;;
    h)  helpMenu
        stop 0
        ;;
    m)  BuildMode="$OPTARG"
        ;;
    r)  RaceDetector="-race"
        ;;
    x)  BuildHash="$OPTARG"
        ;;
    \?) helpMenu
        stop 1
        ;;
    esac
done

#   Assert all of the required arguments are set here
argSet "$BuildMode" "-m"
argSet "$BuildHash" "-x"

#   Other argument validation here...
BuildMode=$(echo "$BuildMode" | tr '[:upper:]' '[:lower:]')

if [[ ! "$BuildMode" =~ "$BuildModeDebug" ]] && [[ ! "$BuildMode" =~ "$BuildModeProduction" ]]; then
    echo -e $RED"ERROR: Invalid build mode \"$BuildMode\" - Must be either \"$BuildModeProduction\" or \"$BuildModeDebug\"."$WHITE
    stop 1
fi

#   Call main, running the full logic of the script.
main

stop 0
