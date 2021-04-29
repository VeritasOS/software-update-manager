#!/bin/bash
# Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9
# Generate an rpm
#
# This template script is used to build RPM from a spec file and inst file.
# Read the comments regarding the section on PLATFORM_RPM_LOCATION below,
# and uncomment it as necessary.
#
# RPMs built are only signed when building on a AS CI Jenkins builders,
# and environment variable, ENABLE_RPM_SIGN=true.  Normally only "release"
# builds have ENABLE_RPM_SIGN=true.
#
# For more details see https://confluence.community.veritas.com/display/ASCD/RPM+Build+Framework
#
# NOTE: Run rpmlint on your rpm before checkin.
#
LICENSE="Copyright (c) `date -u +%Y` Veritas Technologies LLC. All rights reserved.";
LOCKDIR=/tmp/lock.mkrpm
PIDFILE="${LOCKDIR}/PID"
WAIT_INTERVAL=10
RETRY=60
lock_succeed_flag="false"
curdir="$(pwd)"
rpmmacros="$HOME/.rpmmacros"
destdir="${curdir}"
rpmdir="${curdir}/rpmbuild"
topdir="${rpmdir}"
builddir="${topdir}/BUILD"
tmppath="${rpmdir}/tmp"
rpmbuild="/usr/bin/rpmbuild"
coverage_dir="/opt/coverage"
mode=""
enable_sign_rpm="false"
enable_move_rpm="false"
signing_key="appliance.security"
vendor="Veritas Technologies LLC"
packager="Veritas Technologies LLC"
usage="Usage: $0 [-c] -[m] [-s] [-v <version string>] PATH_TO_SPECFILE
    -c Create the coverage specfile based on the specfile provided.
    -m Move rpm to pwd/PRODUCT/BRANCH/BUILDTAG dir, where PRODUCT, BRANCH,
       BUILDTAG are defined by environment variables.  Or move rpm to value
       defined by PLATFORM_RPM_LOCATION.  If PLATFORM_RPM_LOCATION is defined
       by environment variable, then its value will take precedence.
    -s Sign rpm; the default behavior is to NOT sign the RPM, unless this
       switch is active or the env variable ENABLE_RPM_SIGN=true.  Signing
       is only available when running on AS CI Jenkins build servers.
    -v <version string> Provide package version (not release) as parameter.
"

# exit codes and text
MKRPM_SUCCESS=0; ETXT[0]="MKRPM_SUCCESS"
MKRPM_GENERAL=1; ETXT[1]="MKRPM_GENERAL"
MKRPM_LOCKFAIL=2; ETXT[2]="MKRPM_LOCKFAIL"
MKRPM_RECVSIG=3; ETXT[3]="MKRPM_RECVSIG"

# Need lock to ensure that no other mkrpm build is going on.
# mkrpm and rpmbuild should be sigletons, since rpmmacros file is global
# and can be changed at anytime.
for ((i=1;i<="${RETRY}";i++)); do
  if mkdir "${LOCKDIR}"; then
    # lock succeeded, install signal handlers before storing the PID just in case storing the PID fails
    trap 'MKRPMCODE=$?;
          echo "INFO: [mkrpm] $1 - Removing lock. Exit: ${ETXT[MKRPMCODE]}($MKRPMCODE)" >&2
          rm -rf "${LOCKDIR}"' 0

    # the following handler will exit the script upon receiving these signals
    # the trap on "0" (EXIT) from above will be triggered by this script's normal exit
    trap 'echo "INFO:: [mkdir] $1 - Killed by a signal - $1" >&2
          exit ${MKRPM_RECVSIG}' 1 2 3 15

    echo "INFO: [mkrpm] $1 - Installed signal handlers"
    echo "$$" > "${PIDFILE}"
    lock_succeed_flag="true"
    echo "INFO: [mkrpm] $1 - Lock succeeded"
    break
  else
  # lock failed, check if the other PID is alive
    OTHERPID="$(cat "${PIDFILE}")"
    # if cat isn't able to read the file, another instance is probably
    # about to remove the lock
    if [[ $? != 0 ]]; then
      echo "INFO: [mkrpm] $1 - Lock failed, PID ${OTHERPID} is active" >&2
    fi

    if ! kill -0 "${OTHERPID}" &>/dev/null; then
      # lock is stale, remove it and restart
      echo "INFO: [mkrpm] $1 - Removing stale lock of nonexistant PID ${OTHERPID}" >&2
      rm -rf "${LOCKDIR}"
    else
      # lock is valid and OTHERPID is active
      echo "INFO: [mkrpm] $1 - Lock failed, PID ${OTHERPID} is active" >&2
    fi
    echo "INFO: [mkrpm] $1 - Lock failed, waiting $WAIT_INTERVAL sec [retry #$i]"
    sleep "${WAIT_INTERVAL}"
  fi
done
if [[ "${lock_succeed_flag}" != "true" ]]; then
  echo "ERROR: [mkrpm] $1 - Failed to acquire lock after ${RETRY} tries" && exit 1
fi

function err {
    echo "$1"
    exit 1
}

if [[  ! $# -ge 1 ]]; then
    err "${usage}"
fi

# If environment variable ENABLE_RPM_SIGN is TRUE or true, then sign rpm
if [[ "${ENABLE_RPM_SIGN,,}" == "true" ]]; then
    enable_sign_rpm="true"
fi

while getopts ":cmsv:" opt; do
  case ${opt} in
    c) mode="coverage"
       ;;
    m) enable_move_rpm="true"
       ;;
    s) enable_sign_rpm="true"
       ;;
    v) version="${OPTARG}"
       echo "Version passed as param: ${version}"
       ;;
    \?)
       err "${usage}"
       ;;
  esac
done
shift $((OPTIND-1))

specfile="$1"

instfile="${specfile%/*}/$(basename "${specfile}" .spec).inst"
if [[ "${instfile:0:1}" != "/" ]]; then
    instfile="${curdir}/${instfile}"
fi

if [[ ! -e "${specfile}" ]]; then
    err "ERROR: [mkrpm] ${specfile} - ${specfile} spec file not found"
fi

if [[ ! -e "${instfile}" ]]; then
    err "ERROR: [mkrpm] ${specfile} - ${instfile} inst file not found"
fi

# When building from Jenkins CI, use the ${PRODUCT}, ${BRANCH}, and ${BUILDTAG}
# for the destdir and part of the ${BUILDTAG} for the RPM release version
function get_prod_env {
    echo "INFO: [mkrpm] ${specfile} - Get the release version"
    if [[ -z "${PRODUCT}" ]] || [[ -z "${BRANCH}" ]] || [[ -z "${BUILDTAG}" ]]; then
        echo "WARNING: [mkrpm] ${specfile} -  PRODUCT, BRANCH, or BUILDTAG undefined"
        [[ -z "${PRODUCT}" ]] && PRODUCT="${name}"
        [[ -z "${BRANCH}" ]] && BRANCH="main"
        [[ -z "${BUILDTAG}" ]] && BUILDTAG="${version}-${release}"
    fi

    destdir="${curdir}/${PRODUCT}/${BRANCH}/${BUILDTAG}"
    release=$(echo "${BUILDTAG}" | awk -F- '{print $2}')
}

function create_source_archive {
  echo "INFO: [mkrpm] ${specfile} - Create the source archive ${tarfile}"
  tar -zcvf "${tarfile}" $(grep -v ^# "${instfile}" | sed -e "s/\${version}/${version}/g") \
    || err "ERROR: [mkrpm] ${specfile} - Failed to create ${tarfile}"
  mkdir -p "${tmppath}/${name}-${version}"
  (cd "${tmppath}/${name}-${version}"; tar zxvf "${tarfile}")
  (cd "${tmppath}"; tar zcvf "${tarfile}" .)
  rm -rf "${tmppath}/${name}-${version}"
}

function create_rpm_env {
  echo "INFO: [mkrpm] ${specfile} - Create the rpm environment under ${rpmdir}"
  mkdir -p "${rpmdir}"/{RPMS,SRPMS,BUILD,SOURCES,SPECS,tmp}
  echo "INFO: [mkrpm] ${specfile} - For debug purpose cat content of ${rpmmacros}:"
  [[ -e ${rpmmacros} ]] && cat "${rpmmacros}"
  rm -f "${rpmmacros}"
  echo "%packager ${packager}" > "${rpmmacros}"
  specfile_dest="${topdir}/SPECS/${specfile##*/}"
  sed -r -e "s/^Release\s*:\s*.*$/Release: ${release}/" "${specfile}" > "${specfile_dest}" \
    || err "ERROR: [mkrpm] ${specfile} - Copy ${specfile} to ${specfile_dest} failed"
}

function build_rpm {
  echo "INFO: [mkrpm] ${specfile} - Build the ${name} rpm"
  pushd "${rpmdir}" || err "ERROR: [mkrpm] ${specfile} - Can not find ${rpmdir}"
  # Define topdir and tmpdir at the command line to override ~/.rpmmacros
  ${rpmbuild} --define "_topdir ${topdir}" \
              --define "_builddir ${builddir}" \
              --define "_tmppath ${tmppath}" \
              --define "_version ${version}" \
              --define "license ${LICENSE}" \
              --define "vendor ${vendor}" \
              --define "packager ${packager}" \
              -bb "SPECS/${specfile_dest##*/}" \
      || err "ERROR: [mkrpm] ${specfile} - rpmbuild failed to create rpm"
  echo "INFO: [mkrpm] ${specfile} - For debug purpose cat content of ${rpmmacros}:"
  [[ -e ${rpmmacros} ]] && cat "${rpmmacros}"
  popd
}

# Takes as arguments, a list of paths to RPM files
function sign_rpm {
  for _a_rpm in "$@"; do
    echo "INFO: [mkrpm] ${specfile} - Signing ${_a_rpm} with ${signing_key} key"
    if [[ -e "${_a_rpm}" ]]; then
      local _tmp_unsigned_rpm="${_a_rpm%/*}/unsigned_${_a_rpm##*/}"
      mv "${_a_rpm}" "${_tmp_unsigned_rpm}"
      cisigul -v sign-rpm -o "${_a_rpm}" "${signing_key}"  "${_tmp_unsigned_rpm}" ||
        err "ERROR: [mkrpm] ${specfile} - Failed to sign ${_tmp_unsigned_rpm}"
      chmod a+r "${_a_rpm}" && rm -f "${_tmp_unsigned_rpm}"
    else
      err "ERROR: [mkrpm] ${specfile} - Can not find ${_a_rpm} to sign"
    fi
  done
}

# Move rpm for upload to dir so that it can be uploaded artifactory
function move_rpm {
    [[ -n "${PLATFORM_RPM_LOCATION}" ]] && destdir="${PLATFORM_RPM_LOCATION}"
    echo "INFO: [mkrpm] ${specfile} - Move ${rpmfile} to ${destdir}"
    mkdir -p "${destdir}"
    mv "${rpmdir}/RPMS/${buildarch}"/*.rpm "${destdir}/" || \
        err "ERROR: [mkrpm] ${specfile} - Fail to move ${rpmfile} to ${destdir}"
}

# Generate coverage spec and inst files that will be used
# to create the rpm.  These are based on the default spec
# and inst files with added entries for clover db files.
function gen_coverage_spec {
    _specfile_dir="$(dirname ${specfile})"
    _instfile_dir="$(dirname ${instfile})"
    _project_name=$(basename ${specfile} .spec)
    _cov_suffix="_code_coverage"
    _cov_db_list="${_specfile_dir}/${_project_name}${_cov_suffix}_db.lst"

    # Set global variables to the generated coverage spec and inst files
    _orig_specfile="${specfile}"
    _orig_instfile="${instfile}"
    specfile="${_specfile_dir}/${_project_name}${_cov_suffix}.spec"
    instfile="${_instfile_dir}/${_project_name}${_cov_suffix}.inst"

    # Get coverage db files
    readarray clover_dbs < ${_cov_db_list}
    inst_cmds=()
    for db in ${clover_dbs[@]}; do
        inst_cmds+="install -D -m 0600 ${db} %{buildroot}${coverage_dir}/$(basename ${db})\n"
    done

    echo "Generating coverage ${specfile} ..."
    sed -e "/%install/ {
# skip the first line
    n
# append db files
    a ${inst_cmds[@]}
}" -e "/%files/ {
# skip the first line
    n
# append db files
    a %attr(400,-,-) ${coverage_dir}/*
}" "${_orig_specfile}" > "${specfile}"

    echo "INFO: [mkrpm] ${specfile} - Generating coverage ${instfile} ..."
    cat "${_orig_instfile}" <(printf '%s\n' ${clover_dbs[@]}) > "${instfile}"
}

#------------------------------------------------ main

# Generate the coverage spec file
if [[ "${mode}" == "coverage" ]]; then
    gen_coverage_spec
fi

name=$(grep -E "^Name\s*:" "${specfile}" | sed -r -e "s/^Name\s*:\s*(.*)\s*$/\1/")
[[ -z "${version}" ]] && \
    version=$(grep -E "^Version\s*:" "${specfile}" | sed -r -e "s/^Version\s*:\s*(.*)\s*$/\1/")
buildarch=$(grep -E "^BuildArch\s*:" "${specfile}" | sed -r -e "s/^BuildArch\s*:\s*(.*)\s*$/\1/")
release=$(grep -E "^Release\s*:" "${specfile}" | sed -r -e "s/^Release\s*:\s*(.*)\s*$/\1/")

# Only call get_prod_env once name and version (above) has been set
[[ -n "${JENKINS_URL}" ]] && get_prod_env

tarfile="${rpmdir}/SOURCES/${name}-${version}.tar.gz"
rpmfile="${rpmdir}/RPMS/${buildarch}/${name}-${version}-${release}.${buildarch}.rpm"

create_rpm_env
create_source_archive
build_rpm

if [[ -n "${JENKINS_URL}" && "${enable_sign_rpm}" == "true" ]]; then
    sign_rpm "${rpmdir}/RPMS/${buildarch}"/*.rpm
fi

[[ "${enable_move_rpm}" == "true" ]] && move_rpm

echo "INFO: [mkrpm] ${specfile} - Done"
