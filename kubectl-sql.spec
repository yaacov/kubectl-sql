%global provider        github
%global provider_tld    com
%global project         yaacov
%global repo            kubectl-sql
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}

%undefine _missing_build_ids_terminate_build
%define debug_package %{nil}

Name:           %{repo}
Version:        0.3.25
Release:        1%{?dist}
Summary:        kubectl-sql uses sql like language to query the Kubernetes cluster manager
License:        Apache
URL:            https://%{import_path}
Source0:        https://github.com/yaacov/kubectl-sql/archive/v%{version}.tar.gz

BuildRequires:  git
BuildRequires:  golang >= 1.23.0

%description
kubectl-sql let you select Kubernetes resources based on the value of one or more resource fields, using human readable easy to use SQL like query langauge.

%prep
%setup -q -n kubectl-sql-%{version}

%build
# set up temporary build gopath, and put our directory there
mkdir -p ./_build/src/github.com/yaacov
ln -s $(pwd) ./_build/src/github.com/yaacov/kubectl-sql

VERSION=v%{version} make

%install
install -d %{buildroot}%{_bindir}
install -p -m 0755 ./kubectl-sql %{buildroot}%{_bindir}/kubectl-sql

%files
%defattr(-,root,root,-)
%doc LICENSE README.md
%{_bindir}/kubectl-sql

%changelog

* Wed Feb 19 2025 Yaacov Zamir <kobi.zamir@gmail.com> 0.3.16-1
- use TSL v6

* Sun Feb 16 2025 Yaacov Zamir <kobi.zamir@gmail.com> 0.3.14-1
- use TSL v6

* Mon Mar 9 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.11-1
- version should start with v

* Mon Mar 9 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.10-1
- dont show usage on errors

* Mon Mar 9 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.9-1
- preety print join

* Sun Mar 8 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.8-1
- fix docs

* Fri Mar 6 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.6-1
- fix none namespaced resource display

* Fri Mar 6 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.5-1
- add join command

* Thu Mar 5 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.4-1
- rename to kubectl-sql

* Thu Mar 4 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.2-1
- use git version number

* Thu Mar 4 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.1-1
- Fix multiple resources

* Thu Mar 4 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.2.0-1
- Use kubectl plugin kit

* Thu Feb 22 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.1.18-1
- Fix float printing

* Thu Feb 22 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.1.17-1
- Add config option

* Thu Feb 22 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.1.16-1
- Fix parsing of anotations

* Thu Feb 22 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.1.15-1
- Fix parsing of labels and anotations

* Thu Feb 22 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.1.14-1
- Fix parsing of numbers

* Thu Feb 22 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.1.13-1
- Parse dates and booleans

* Thu Feb 20 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.1.12-1
- No debug rpm

* Thu Feb 20 2020 Yaacov Zamir <kobi.zamir@gmail.com> 0.1.11-1
- Initial RPM release
