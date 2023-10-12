Name:           svipul
Version:        0.1.06g398fa78dirty
Release:        1
Summary:        Svipul SNMP collector

Group:          telenornms
License:        LGPL-2.1
URL:            https://github.com/telenornms/svipul
Source0:        https://github.com/telenornms/svipul/archive/v0.1.0-6-g398fa78-dirty.tar.gz


BuildArch:      x86_64
# Since we download go manually and not through yum,
# the version won't be registered as installed.
#BuildRequires:  go >= 1.13
#Debian hack, auto-commented out: BuildRequires:  python3-docutils, systemd-units


%description
Svipul is a tool for collecting SNMP data based on orders received over
RabbitMQ.

# Executable files require a build id; let's stop that
# https://github.com/rpm-software-management/rpm/issues/367
%undefine _missing_build_ids_terminate_build

%prep
%setup -q

%build
make

%install
make install DESTDIR=%{buildroot} PREFIX=/usr DOCDIR=%{_defaultdocdir}/svipul-%{version}
install -D -m 0644 build/%{name}.service %{buildroot}%{_unitdir}/%{name}.service

%pre
getent group svipul >/dev/null || groupadd -r svipul
getent passwd svipul >/dev/null || \
       useradd -r -g svipul -d /var/lib/svipul -s /sbin/nologin \
               -c "Svipul metric collector" svipul
exit 0

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service


%files
%license LICENSE
%{_bindir}/%{name}
#%{_mandir}/man1/%{name}.1*
%docdir %{_defaultdocdir}/%{name}-%{version}
%{_defaultdocdir}/%{name}-%{version}
%{_unitdir}/%{name}.service
%config %{_sysconfdir}/%{name}/output.d/default.json



%changelog
