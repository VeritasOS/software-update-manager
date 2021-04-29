# Copyright (c) 2020 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

Name:           __NAME__
Version:        __VERSION__
Release:        __RELEASE__
Summary:        __SUMMARY__

License:        %{license}
URL:            __URL__
SOURCE0:        %{name}-%{version}.tar.gz

BuildArch:      x86_64
Packager:       Appliance Solutions

%description
ASUM RPM Format Version : __ASUM_RPM_FORMAT_VERSION__
RPM Info    : __RPM_INFO__

%define debug_package %{nil}

%prep
%setup -q

%build

%install
rm -rf %{buildroot}

install -m 0755 -d %{buildroot}/
cp -rp  ./* %{buildroot}/

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
/*

%pre
__PRESCRIPT__

%post

%postun

%changelog