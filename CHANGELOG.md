# Changelog

## [6.0.0](https://github.com/hasrthur/agent-sandbox.nix/compare/v5.10.0...v6.0.0) (2026-11-27)


### ⚠ BREAKING CHANGES

* rename allowedLocalPorts -> allowedHostPorts
* Linux sandboxes that implicitly relied on AF_UNIX sockets must set allowUnixSockets = true.
* migrate to python
* declared rwDirs/rwFiles must exist before launch
* fail closed on git identity instead of fabricating one
* Renamed extraEnv → env. Pure rename; semantics unchanged.

### darwin

* allow AF_UNIX bind/connect scoped to writable dirs (CWD + rwD… ([#92](https://github.com/hasrthur/agent-sandbox.nix/issues/92)) ([dbca506](https://github.com/hasrthur/agent-sandbox.nix/commit/dbca50647f4e9f5933b3c573ee7b6e4552526129))


### Features

* add _proxyRedirects private arg to sandbox ([a1000ce](https://github.com/hasrthur/agent-sandbox.nix/commit/a1000ce130f03d47f5e00eeaf93f7f54437b6a3e))
* add /bin/sh to path ([b5989f8](https://github.com/hasrthur/agent-sandbox.nix/commit/b5989f89b814648e3cb075a262ee47dc7f5e56e3))
* add claude-nix shell.nix ([76dae03](https://github.com/hasrthur/agent-sandbox.nix/commit/76dae037721b1e9fc83f483631fa84af60acc34d))
* allow launching from home ([#77](https://github.com/hasrthur/agent-sandbox.nix/issues/77)) ([d147b38](https://github.com/hasrthur/agent-sandbox.nix/commit/d147b38fe62a74f70036aa9203df3b7e8956375d))
* allowedInboundPorts — host-to-sandbox TCP port forwards ([#116](https://github.com/hasrthur/agent-sandbox.nix/issues/116)) ([c84478a](https://github.com/hasrthur/agent-sandbox.nix/commit/c84478ae3ccd0f6fbbc790ef2ce6907e9e4390db))
* allowHeadlessBrowsers ([#179](https://github.com/hasrthur/agent-sandbox.nix/issues/179)) ([bfad3a1](https://github.com/hasrthur/agent-sandbox.nix/commit/bfad3a1818d36c62257e0bac3c4b3852b8b7b5c2))
* **allowLocalPorts:** Add cross-platform localNetworkAccess targets ([#68](https://github.com/hasrthur/agent-sandbox.nix/issues/68)) ([3a0930a](https://github.com/hasrthur/agent-sandbox.nix/commit/3a0930a2dc9e24d9f93d4c7998f36b1a0a5d272d))
* **allowNix:** refuse a trusted user, confirm an unsandboxed daemon ([#163](https://github.com/hasrthur/agent-sandbox.nix/issues/163)) ([0349ba0](https://github.com/hasrthur/agent-sandbox.nix/commit/0349ba0055899b6a39f92ee26c296a1d08603a9a))
* antigravity-cli template ([#209](https://github.com/hasrthur/agent-sandbox.nix/issues/209)) ([20e534a](https://github.com/hasrthur/agent-sandbox.nix/commit/20e534ab2640b1f8fd96fe8a679339b0bd4e19bc))
* bash non interactive warning on darwin ([2c187f6](https://github.com/hasrthur/agent-sandbox.nix/commit/2c187f62f2d646f7d7a528267d9f6123ef877923))
* bind symlinked state files on linux ([f2c39b9](https://github.com/hasrthur/agent-sandbox.nix/commit/f2c39b9d8b898d21658f3a2b9dba61a72e19c5cb))
* bindRepoRoot param ([0bda54c](https://github.com/hasrthur/agent-sandbox.nix/commit/0bda54c1c784fbeafe5dc4fe3db17a79ba8a6fb6))
* block dns resolution on linux when restrictNetwork=true ([fa26718](https://github.com/hasrthur/agent-sandbox.nix/commit/fa26718aee3186c60f492ea84676e19931766da6))
* CI for PRs into main ([45a5a50](https://github.com/hasrthur/agent-sandbox.nix/commit/45a5a5070603e93aeed90dae1074818715f4d607))
* ci on PRs to main ([51d1323](https://github.com/hasrthur/agent-sandbox.nix/commit/51d13239300fd51c5fcff084d466b45ad1f4c860))
* **commonTools:** add curl ([e9cd55d](https://github.com/hasrthur/agent-sandbox.nix/commit/e9cd55db3bef751f41de02ea43a16bb3bebbd909))
* **darwin:** allowNix ([6cea43e](https://github.com/hasrthur/agent-sandbox.nix/commit/6cea43e701186fd8d636bf39f871f06a1f5af0bc))
* debug steps ([85b7e7e](https://github.com/hasrthur/agent-sandbox.nix/commit/85b7e7e19f9010f23f05a2f74b687143d2136879))
* declared rwDirs/rwFiles must exist before launch ([1342f80](https://github.com/hasrthur/agent-sandbox.nix/commit/1342f808651dd3fe71b28a033158dd76b1df0117))
* fail closed on git identity instead of fabricating one ([f1122d1](https://github.com/hasrthur/agent-sandbox.nix/commit/f1122d1b920831ee6463420eefd8b8206f46a14e))
* go proxy ([10d8ecb](https://github.com/hasrthur/agent-sandbox.nix/commit/10d8ecbd6968771ee3d9b77aac137521333be37f))
* injectedCredentials ([#207](https://github.com/hasrthur/agent-sandbox.nix/issues/207)) ([dcab096](https://github.com/hasrthur/agent-sandbox.nix/commit/dcab0968a0946a063981e7f6c98b41a102238352))
* license ([4d5b41e](https://github.com/hasrthur/agent-sandbox.nix/commit/4d5b41eced852c916b135ba7217576e6e0887dec))
* **linux:** allowNix ([9f0219d](https://github.com/hasrthur/agent-sandbox.nix/commit/9f0219da1156a206573dcf5ddd117bf10c43aee5))
* log injected credentials ([#211](https://github.com/hasrthur/agent-sandbox.nix/issues/211)) ([277aeef](https://github.com/hasrthur/agent-sandbox.nix/commit/277aeefb7b566d02cc6c31b5a735d6e610e556ec))
* make bash required ([a7d1399](https://github.com/hasrthur/agent-sandbox.nix/commit/a7d1399a726008afb0d3281456af5f2dc6f499d0))
* migrate to python ([7695a5f](https://github.com/hasrthur/agent-sandbox.nix/commit/7695a5f6319275aa75a816787424f57fdd83bfae))
* network restrictions ([7dd28bf](https://github.com/hasrthur/agent-sandbox.nix/commit/7dd28bf695aee4b4b45df8ace075299dd9b7d332))
* print version and session at launch ([#188](https://github.com/hasrthur/agent-sandbox.nix/issues/188)) ([3cb1e72](https://github.com/hasrthur/agent-sandbox.nix/commit/3cb1e72cf9873b9f2ab4f759fea41be1eff4d22b))
* **proxy:** allow per-path rules for allowedDomains ([#202](https://github.com/hasrthur/agent-sandbox.nix/issues/202)) ([dbc373b](https://github.com/hasrthur/agent-sandbox.nix/commit/dbc373b49edf2df48626e9360f33e4e76290f4c6))
* **proxy:** log each allowed host on first contact ([#123](https://github.com/hasrthur/agent-sandbox.nix/issues/123)) ([a69b1ee](https://github.com/hasrthur/agent-sandbox.nix/commit/a69b1eed7e15cede8ba52f90e80df2c6e355ff7c))
* **proxy:** websockets ([#127](https://github.com/hasrthur/agent-sandbox.nix/issues/127)) ([4f1b185](https://github.com/hasrthur/agent-sandbox.nix/commit/4f1b1854f3048c8e0145a8b3c3e847af98214d55))
* **README:** allowNix ([bcea1b2](https://github.com/hasrthur/agent-sandbox.nix/commit/bcea1b28f4ed2c4aff2d2672152d6963f0e5604d))
* relicense under MIT ([1462f69](https://github.com/hasrthur/agent-sandbox.nix/commit/1462f69154f33efd6b809ed1a598dd405673128c))
* rename allowedLocalPorts -&gt; allowedHostPorts ([99b75c8](https://github.com/hasrthur/agent-sandbox.nix/commit/99b75c820c9a2b6055250f4de6eb78a7641f855b))
* rename API args and replace restrictNetwork with allowedDomains ([a2ee921](https://github.com/hasrthur/agent-sandbox.nix/commit/a2ee921d1ff2b158d8391fb5f22ca5774d5955f8))
* restrict nix/store access to allowedPackages ([2ef361a](https://github.com/hasrthur/agent-sandbox.nix/commit/2ef361a87c3919055dcd4bb999fd994c9df342ed))
* roDirs and roFiles read-only bind primitives ([54066d0](https://github.com/hasrthur/agent-sandbox.nix/commit/54066d013545451f1bda2974497ef000c4ae2608))
* roDirs and roFiles read-only bind primitives ([5f8ce44](https://github.com/hasrthur/agent-sandbox.nix/commit/5f8ce44c26c0cfca299f9ac9c534dfe9d1be4773))
* startup log ([#97](https://github.com/hasrthur/agent-sandbox.nix/issues/97)) ([9451d86](https://github.com/hasrthur/agent-sandbox.nix/commit/9451d86e545490c162bc9b2bbea2254a06f4d0c4))
* templates ([0fea2e9](https://github.com/hasrthur/agent-sandbox.nix/commit/0fea2e99e120fd9a31d9f032b22d0217b7bd5d5f))
* templates ([0d53dfd](https://github.com/hasrthur/agent-sandbox.nix/commit/0d53dfdf4d175b26ac2e382c274f7c96916d06a3))
* tests and CI ([8be5e02](https://github.com/hasrthur/agent-sandbox.nix/commit/8be5e02e082c5a6a396f074224bdf93f0bd74250))
* timestampt network logs ([4b4f14f](https://github.com/hasrthur/agent-sandbox.nix/commit/4b4f14fd567eda14e5915c6f79f9b60364ce3edd))
* TODO.md ([8543e1e](https://github.com/hasrthur/agent-sandbox.nix/commit/8543e1eb37e446259cecc473f4f4ea9d8d0d4062))
* update README.md ([83f018a](https://github.com/hasrthur/agent-sandbox.nix/commit/83f018aeaf09c9ad091d310285a4b537ec0e5892))
* use ephemeral HOME on macOS ([f6c7ac5](https://github.com/hasrthur/agent-sandbox.nix/commit/f6c7ac58338752ea9fb567813566ff1db210d2f1))
* version ([#99](https://github.com/hasrthur/agent-sandbox.nix/issues/99)) ([b2217b6](https://github.com/hasrthur/agent-sandbox.nix/commit/b2217b6f91e050608cc5a9f1c75eab54f5859103))
* workspaceDir ([#169](https://github.com/hasrthur/agent-sandbox.nix/issues/169)) ([ee965e5](https://github.com/hasrthur/agent-sandbox.nix/commit/ee965e584f5a94d374ea4e8ea6ee47187ea4e79e))
* write proxy port to fifo ([e1a9202](https://github.com/hasrthur/agent-sandbox.nix/commit/e1a92023aa35a6dc4e15122c2305302dd4d255f1))
* write proxy port to fifo ([5ee023e](https://github.com/hasrthur/agent-sandbox.nix/commit/5ee023e3f8a98fd96f65dc8da46fbec316cd1094))


### Bug Fixes

* add cacert to closure ([41199ee](https://github.com/hasrthur/agent-sandbox.nix/commit/41199eec4867e79878995d62e76a39f8a10b9c8f))
* add intermediate symlinks to bubblewrap sandbox ([95d80d8](https://github.com/hasrthur/agent-sandbox.nix/commit/95d80d8b539508afdcd0ead5cacbda967b5f37b1))
* add retries to curl tests ([1773656](https://github.com/hasrthur/agent-sandbox.nix/commit/1773656f417e213850f34ddc213351fc7d80866a))
* **allowNix:** ensure daemon ([#135](https://github.com/hasrthur/agent-sandbox.nix/issues/135)) ([8fa9e44](https://github.com/hasrthur/agent-sandbox.nix/commit/8fa9e44e12eefd1bcd63a663c8ed40eec26f6ef1))
* block TCP to pasta gateway on non-proxy ports ([84828d6](https://github.com/hasrthur/agent-sandbox.nix/commit/84828d6d04ff1cc660bf91cdd55b69f31ab38b67))
* CLAUDE_CONFIG_DIR update ([3a701d8](https://github.com/hasrthur/agent-sandbox.nix/commit/3a701d86d93cc58572e181e92ea15fa9ff5575d8))
* **claude:** add .claude.json.lock ([eeba61b](https://github.com/hasrthur/agent-sandbox.nix/commit/eeba61bc4bee4db503db3ac484d9090e97097218))
* correct AUDIT_ARCH_AARCH64 in the AF_UNIX seccomp filter ([21d13be](https://github.com/hasrthur/agent-sandbox.nix/commit/21d13bec3deeb4c427fa0d6c98f97ef4b4e37638))
* **darwin:** add the session tmpdir to the AF_UNIX socket scope ([#121](https://github.com/hasrthur/agent-sandbox.nix/issues/121)) ([b8c1e45](https://github.com/hasrthur/agent-sandbox.nix/commit/b8c1e459dc9b4765f403e4e1437797a335f0bc1f))
* **darwin:** allow subshell exec ([6d8b239](https://github.com/hasrthur/agent-sandbox.nix/commit/6d8b239b7bd57f5d27366ef63eec6ebfc96df0be))
* **darwin:** block procargs reads of host processes ([#183](https://github.com/hasrthur/agent-sandbox.nix/issues/183)) ([9253163](https://github.com/hasrthur/agent-sandbox.nix/commit/925316356e501ccf3a90f427d4df70c302a0fd92))
* **darwin:** deny kern.proc* sysctls ([436436a](https://github.com/hasrthur/agent-sandbox.nix/commit/436436a69adf422883d516bdde0e6c9c1e5e6225))
* **darwin:** deny process info instead of starting a new session ([#204](https://github.com/hasrthur/agent-sandbox.nix/issues/204)) ([d78cfcd](https://github.com/hasrthur/agent-sandbox.nix/commit/d78cfcd59a99f382bcad9dcf73b530a4751facd0))
* **darwin:** deny writes to read-only paths inside writable ones ([#140](https://github.com/hasrthur/agent-sandbox.nix/issues/140)) ([461ea4b](https://github.com/hasrthur/agent-sandbox.nix/commit/461ea4b0e7a6559d4f380b07bf9b03c8f725388d))
* **darwin:** don't load bash profile ([8673940](https://github.com/hasrthur/agent-sandbox.nix/commit/867394086f4de95317c41c8270200264f07e1a4f))
* **darwin:** dummy private/etc/passwwd ([7bfa770](https://github.com/hasrthur/agent-sandbox.nix/commit/7bfa770296d7722a35fa4893f6a2a459dd7a0e2d))
* **darwin:** file-read-metadata not file-read* for traversal ([f7fb67b](https://github.com/hasrthur/agent-sandbox.nix/commit/f7fb67be3d600c47de73320fdb894e99899b7494))
* **darwin:** file-read-metadata not file-read* for traversal ([5fc2886](https://github.com/hasrthur/agent-sandbox.nix/commit/5fc2886008d162ab25596987011f587e0b513640))
* **darwin:** file-read-metadata traversal + process-exec sandbox HOME ([9d35012](https://github.com/hasrthur/agent-sandbox.nix/commit/9d35012e9bdec268388d0a47988c50d970ef2e47))
* **Darwin:** fix whitespace string issue for allowedLocalPort seatbelt str ([637511e](https://github.com/hasrthur/agent-sandbox.nix/commit/637511e5fc7613804e6a4643eea9c73494e35c53))
* **darwin:** follow symlink chains ([#160](https://github.com/hasrthur/agent-sandbox.nix/issues/160)) ([4664d0b](https://github.com/hasrthur/agent-sandbox.nix/commit/4664d0b438c37b99628ae1de837ca18bd82fc0a5))
* **darwin:** forward terminal resizes to the sandbox ([#192](https://github.com/hasrthur/agent-sandbox.nix/issues/192)) ([1e09e34](https://github.com/hasrthur/agent-sandbox.nix/commit/1e09e348a5b3304801bae0d97290ff972fe3ab1e))
* **darwin:** grant file-read-metadata on ancestor dirs between HOME and CWD ([bd10d4f](https://github.com/hasrthur/agent-sandbox.nix/commit/bd10d4f1f0fb2d57261e11bc7996b2b9c5dbf10a))
* **darwin:** let nix 2.34.9 stat the Rosetta runtime ([42ce94b](https://github.com/hasrthur/agent-sandbox.nix/commit/42ce94b657e85dae6a6d5f30d4ee0ee70399ce9f))
* **darwin:** let the sandbox control its new terminal ([8e1e563](https://github.com/hasrthur/agent-sandbox.nix/commit/8e1e563ca1881d2bb6f0c7606a2b205dbba12827))
* **darwin:** limit outbound localhost to proxy ([90b848b](https://github.com/hasrthur/agent-sandbox.nix/commit/90b848b5d7ccdbc1469970d325f9cbd14ea16fac))
* **darwin:** make .git/hooks and .git/config read-only in sandbox ([ac3616e](https://github.com/hasrthur/agent-sandbox.nix/commit/ac3616ec6732694e2aec0553a28cd55c701517d5))
* **darwin:** max private/var unreachable ([d5d10be](https://github.com/hasrthur/agent-sandbox.nix/commit/d5d10becb74ae4366ce48a82323cb0935eae0dc7))
* **darwin:** process exec plutil ([8083620](https://github.com/hasrthur/agent-sandbox.nix/commit/80836207c84783f73239411662c967e3ed85a719))
* **darwin:** process-exec sandbox home ([e4ee582](https://github.com/hasrthur/agent-sandbox.nix/commit/e4ee58222ce266c11965f5c51993544063240daf))
* **darwin:** process-exec sandbox home ([f7742d1](https://github.com/hasrthur/agent-sandbox.nix/commit/f7742d145236d6250c55e4118c222420337fdeb9))
* **darwin:** refuse binds nested inside another declared bind ([#87](https://github.com/hasrthur/agent-sandbox.nix/issues/87)) ([bd51a8b](https://github.com/hasrthur/agent-sandbox.nix/commit/bd51a8b030c796cbbdbab89c76fab79fe7c26e57))
* **darwin:** resolve real nix daemon socket path ([2e9dd51](https://github.com/hasrthur/agent-sandbox.nix/commit/2e9dd5171b70bb0c727b16f21bdf2351fd285f6c))
* **darwin:** restore resize and Ctrl+Z under the new session ([#194](https://github.com/hasrthur/agent-sandbox.nix/issues/194)) ([407b549](https://github.com/hasrthur/agent-sandbox.nix/commit/407b549e0e1d2440212c5160259a48e26c0af833))
* **darwin:** restrict nix store reads to symlinked files + allowedPackages ([#133](https://github.com/hasrthur/agent-sandbox.nix/issues/133)) ([61bdf31](https://github.com/hasrthur/agent-sandbox.nix/commit/61bdf31ca9426e4f7fb77632e188c172a13fbae9))
* **darwin:** rm /Library/Preferences read + plutil exec ([33ad129](https://github.com/hasrthur/agent-sandbox.nix/commit/33ad129495ccd43e641f77a1038afb88e7edf493))
* **darwin:** rm unix socket access ([74a28c3](https://github.com/hasrthur/agent-sandbox.nix/commit/74a28c3f495ec904e4fa20b5ab8be78d4550d8a9))
* **darwin:** run git hooks when the gitdir is outside the launch directory ([#167](https://github.com/hasrthur/agent-sandbox.nix/issues/167)) ([a70a90c](https://github.com/hasrthur/agent-sandbox.nix/commit/a70a90cb4e730e4edbc49fc6bd76fbaf76b794eb))
* **darwin:** set CLAUDE_CODE_TMPDIR ([#118](https://github.com/hasrthur/agent-sandbox.nix/issues/118)) ([6e517f5](https://github.com/hasrthur/agent-sandbox.nix/commit/6e517f5e2181bcf227675642456560365cccf47d))
* **darwin:** start the sandbox in its own session ([#185](https://github.com/hasrthur/agent-sandbox.nix/issues/185)) ([41477e2](https://github.com/hasrthur/agent-sandbox.nix/commit/41477e251b85aca18d3291af73616026a0208435))
* **darwin:** test-unix-socket* - use nix provided python3 ([89abb40](https://github.com/hasrthur/agent-sandbox.nix/commit/89abb40ff8e5ea49d664746acce0ae9747aea101))
* **darwin:** write $HOME and $TMPDIR to session state dir ([#103](https://github.com/hasrthur/agent-sandbox.nix/issues/103)) ([0f7647a](https://github.com/hasrthur/agent-sandbox.nix/commit/0f7647aacf3bf4a9a45b227edcd9192f426590d9))
* **default.nix:** missing ; ([297df91](https://github.com/hasrthur/agent-sandbox.nix/commit/297df919ab7b3d7afbefef41b9e240fe919f845a))
* disable .git discovery when $HOME==$REPO_ROOT ([72ac65c](https://github.com/hasrthur/agent-sandbox.nix/commit/72ac65c108761af326e8403c40a736ae755a6b92))
* ensure etc/hosts exists on linux ([d4d1993](https://github.com/hasrthur/agent-sandbox.nix/commit/d4d19939aff52b3334b3315ff26e5836f8ab074d))
* ensure traversal from cwd to stateDirs and stateFiles ([d91aa0c](https://github.com/hasrthur/agent-sandbox.nix/commit/d91aa0c0685514f6a4bd89de40f3826f98ba131f))
* ensure usr/bin/env on linux ([30e6ce7](https://github.com/hasrthur/agent-sandbox.nix/commit/30e6ce7715ccb67bc1461b062deddaa37f1b6079))
* example / debug shells ([eaebac0](https://github.com/hasrthur/agent-sandbox.nix/commit/eaebac002ca802dee898316a37b607d50f6cc41a))
* **example.shell.nix:** config.allowUnfree=true ([f72a30d](https://github.com/hasrthur/agent-sandbox.nix/commit/f72a30dbb3311fd923f63bd5082bd8069c6f1bad))
* exec froms state dirs ([3ca33b8](https://github.com/hasrthur/agent-sandbox.nix/commit/3ca33b8d2f1eeff2d3e544de403b1b9127e51b1f))
* **git:** grant the work tree root rather than the common git dir's parent ([#102](https://github.com/hasrthur/agent-sandbox.nix/issues/102)) ([bd7a87a](https://github.com/hasrthur/agent-sandbox.nix/commit/bd7a87a5189011adb64ab694ae1014c99ad84b03))
* **git:** pin objects/info/alternates in every gitdir ([#153](https://github.com/hasrthur/agent-sandbox.nix/issues/153)) ([2b6647a](https://github.com/hasrthur/agent-sandbox.nix/commit/2b6647a5e2a04675d30833ed7fe269205aacc50b))
* launch in a subshell ([#172](https://github.com/hasrthur/agent-sandbox.nix/issues/172)) ([542427e](https://github.com/hasrthur/agent-sandbox.nix/commit/542427ebb50039e00e08f8f928dc55921cac259a))
* **launch log:** make bwrap.args readable ([ffd47e0](https://github.com/hasrthur/agent-sandbox.nix/commit/ffd47e0db1b4b4443fef8727f02213932f1bdb86))
* **launcher:** always set SANDBOX_PROXY_REDIRECT ([4a30596](https://github.com/hasrthur/agent-sandbox.nix/commit/4a305962be3b27a7807117e12c9905d947aa7eb5))
* **linux:** bind an empty resolv.conf in restricted mode ([#200](https://github.com/hasrthur/agent-sandbox.nix/issues/200)) ([f90f7af](https://github.com/hasrthur/agent-sandbox.nix/commit/f90f7af92792f9d7aeb0fe159c1c15011fc3dbd3))
* **linux:** bind proxy on 127.0.0.1 and route via pasta gateway to avoid leaking host LAN IP ([15e6725](https://github.com/hasrthur/agent-sandbox.nix/commit/15e67255c95de9987ece24a0449e3c5d25d9f1f8))
* **linux:** bind proxy to host IP instead of 0.0.0.0 ([08bd8a6](https://github.com/hasrthur/agent-sandbox.nix/commit/08bd8a62b4fefee7244c4bf427ba53160091fb63))
* **linux:** bind roFile/rwFile symlinks at their declared paths ([4c9cac0](https://github.com/hasrthur/agent-sandbox.nix/commit/4c9cac00437dbfcdf26cf013e434da53a7954fa2))
* **linux:** clear bwrap environ to prevent host env leak via /proc/1/environ ([1a53cb6](https://github.com/hasrthur/agent-sandbox.nix/commit/1a53cb617ac2b950bd042559d75eeedec81e7194))
* **linux:** disable IPv6 on the sandbox's network interface ([#190](https://github.com/hasrthur/agent-sandbox.nix/issues/190)) ([427b320](https://github.com/hasrthur/agent-sandbox.nix/commit/427b3207c8ddb385b360d1d02a4c6e93fcef0d5c))
* **linux:** don't follow symlinks when binding files ([67e7018](https://github.com/hasrthur/agent-sandbox.nix/commit/67e70185e91eb98983a2dbeea73d870d57ef5477))
* **linux:** dummy private/etc/passwwd ([4362548](https://github.com/hasrthur/agent-sandbox.nix/commit/43625485dd3949866bb0abd9adea1ab4427ce943))
* **linux:** isolate sandbox network from host local services ([c3be4b2](https://github.com/hasrthur/agent-sandbox.nix/commit/c3be4b22ab45d32a2b4a02d78d51fbc5593a4869))
* **linux:** make .git/hooks and .git/config read-only in sandbox ([5e33e9b](https://github.com/hasrthur/agent-sandbox.nix/commit/5e33e9b443c6e6a08ea8989b546b65ec67b1bb37))
* **linux:** make nftables deny default ([9eb1ee2](https://github.com/hasrthur/agent-sandbox.nix/commit/9eb1ee2bdd27d6d73de49134e070a2222d4a4389))
* **linux:** mask /proc/cmdline and /proc/sys/kernel/random/boot_id ([ee715f5](https://github.com/hasrthur/agent-sandbox.nix/commit/ee715f54f4de7ae63d9e04f175930ca7ee2cfc5e))
* **linux:** refuse launch for unresolved symlink hop ([#143](https://github.com/hasrthur/agent-sandbox.nix/issues/143)) ([8b72979](https://github.com/hasrthur/agent-sandbox.nix/commit/8b729792aaf573dd8205d095dc5d7e7fa8252ff4))
* **linux:** set neutral hostname in UTS namespace to prevent host leak ([2d164b7](https://github.com/hasrthur/agent-sandbox.nix/commit/2d164b74b6b82c00fbdd6f8177f4595fe6003690))
* make nix store readable on darwin ([ddbd3c7](https://github.com/hasrthur/agent-sandbox.nix/commit/ddbd3c7422af16809192d7a5e79761efe1c8f7f8))
* make nix store readable on darwin ([82f6c74](https://github.com/hasrthur/agent-sandbox.nix/commit/82f6c74ca8dae3fa68d4185a52bbb1ea70789e5d))
* max-depth=1 for symlink search ([2fb9b6f](https://github.com/hasrthur/agent-sandbox.nix/commit/2fb9b6f10588c98b6ab2459d0e7e0b5059bd9936))
* max-depth=1 for symlink search ([19407a0](https://github.com/hasrthur/agent-sandbox.nix/commit/19407a0dfc5321a5d248e83db6535aab519996da))
* only allow state symlinks to nix store ([2523711](https://github.com/hasrthur/agent-sandbox.nix/commit/25237110a6580ccad25e3cd80163c0c1e605a5be))
* **pasta:** use ipv4 ([f1f55e9](https://github.com/hasrthur/agent-sandbox.nix/commit/f1f55e9702c97b380e7bc41dbf9472b3d006a77d))
* protect the whole repo's gitdir from hook injection ([#79](https://github.com/hasrthur/agent-sandbox.nix/issues/79)) ([147d9af](https://github.com/hasrthur/agent-sandbox.nix/commit/147d9af8c8936c9937767f698d2774a5709ddbe7))
* **proxy:** bound upstream dials and cap concurrent connections ([#147](https://github.com/hasrthur/agent-sandbox.nix/issues/147)) ([baa4ae3](https://github.com/hasrthur/agent-sandbox.nix/commit/baa4ae32a3283edcfb5ac8bc0a406a3561c2505a))
* **proxy:** cap request headers, deadline the client handshake, log accept backoff ([#149](https://github.com/hasrthur/agent-sandbox.nix/issues/149)) ([484524a](https://github.com/hasrthur/agent-sandbox.nix/commit/484524a05dab5b9f120ae1e4df3a2d3a958252bf))
* **proxy:** disallow , and = in redirects ([#146](https://github.com/hasrthur/agent-sandbox.nix/issues/146)) ([f1397bc](https://github.com/hasrthur/agent-sandbox.nix/commit/f1397bcd42363b046c7cde1273be887d1e26a62c))
* **proxy:** disallow addresses resolving to loopback ([#86](https://github.com/hasrthur/agent-sandbox.nix/issues/86)) ([8265aa5](https://github.com/hasrthur/agent-sandbox.nix/commit/8265aa50592d5c3aa3559e7115702bcb8ac399a0))
* **proxy:** document + test default policy ([#152](https://github.com/hasrthur/agent-sandbox.nix/issues/152)) ([b4e36b2](https://github.com/hasrthur/agent-sandbox.nix/commit/b4e36b2effafc2611f1ea0556db60cee1fa0e7f4))
* **proxy:** fold hostnames ASCII-only and refuse a non-ASCII host ([#151](https://github.com/hasrthur/agent-sandbox.nix/issues/151)) ([019780f](https://github.com/hasrthur/agent-sandbox.nix/commit/019780facaf266b3b7e08b4095ff4857f76ef4d0))
* **proxy:** reject non-443 CONNECT and non-80 plaintext requests ([4880cf4](https://github.com/hasrthur/agent-sandbox.nix/commit/4880cf43463c10235da9a74e3d7cf098ab538172))
* **proxy:** require the Host header to match the CONNECT host ([#150](https://github.com/hasrthur/agent-sandbox.nix/issues/150)) ([8813839](https://github.com/hasrthur/agent-sandbox.nix/commit/8813839576a062c169589cf68479e9e835293816))
* **proxy:** write the 101 handshake in a single write ([#129](https://github.com/hasrthur/agent-sandbox.nix/issues/129)) ([e9176ab](https://github.com/hasrthur/agent-sandbox.nix/commit/e9176ab0299d29eb92f6ea5d526088c8bc7f9f90))
* read dirname(/) bug ([f0263e3](https://github.com/hasrthur/agent-sandbox.nix/commit/f0263e3eeb7320020a20b262a166533eabdbccc0))
* README.md ([2786343](https://github.com/hasrthur/agent-sandbox.nix/commit/2786343b2175fd8996215993a56da060b12e2525))
* **README:** macos keychain workaround ([6350246](https://github.com/hasrthur/agent-sandbox.nix/commit/6350246267c94f86ed0f876bb11ef651ef50e510))
* remove inheritPath ([2222a8c](https://github.com/hasrthur/agent-sandbox.nix/commit/2222a8cc11d829a0523c49d33b2c7ce3702204a3))
* remove now redundant nix/store tests ([2070e1d](https://github.com/hasrthur/agent-sandbox.nix/commit/2070e1de7ff533a40f8680cec7722ba27e600165))
* remove read access to real HOME directory in macOS sandbox ([b804a7b](https://github.com/hasrthur/agent-sandbox.nix/commit/b804a7b982e895c292ea063ce8badfdf72dc19ce))
* resolve resolv.conf on ubuntu ([cc2d145](https://github.com/hasrthur/agent-sandbox.nix/commit/cc2d1453b43269dc8c3869e4cee160cb7a1d385c))
* restrict /dev/tty access on macos ([471aae2](https://github.com/hasrthur/agent-sandbox.nix/commit/471aae2d5e7c2e5b5bdd89d7d24711c6e80e10be))
* restrict /dev/tty access on macos ([aaf06d3](https://github.com/hasrthur/agent-sandbox.nix/commit/aaf06d3a6378db22ee23877931a42e5eb3a6bcbf))
* rm .pem suffixes from mktemp calls ([97676d5](https://github.com/hasrthur/agent-sandbox.nix/commit/97676d554e7d03079ac856ac58f43cae8003a5b0))
* rm python dependency in darwin tests ([76af178](https://github.com/hasrthur/agent-sandbox.nix/commit/76af178967dd2caae3423578e954a3373b2b5612))
* **seatbelt:** various additions ([136841e](https://github.com/hasrthur/agent-sandbox.nix/commit/136841e438b74e9d08a74317671819a4099db3c5))
* set PKG_CONFIG_PATH ([#138](https://github.com/hasrthur/agent-sandbox.nix/issues/138)) ([f801173](https://github.com/hasrthur/agent-sandbox.nix/commit/f8011734d69bbe2fdc0c3f568a7bb3ec98535e47))
* symlink test ([89d5d77](https://github.com/hasrthur/agent-sandbox.nix/commit/89d5d7731ae3b1ed01c363639b2a5473a32a4984))
* **symlinks:** only skip seen dirs / files with real content ([287c68a](https://github.com/hasrthur/agent-sandbox.nix/commit/287c68a2ff95cb8428b96feccb10e8e1a1648cd2))
* sync shell version ([7097ca1](https://github.com/hasrthur/agent-sandbox.nix/commit/7097ca13cb2129406373abc56816bae38e3b5573))
* test module rename ([96161f7](https://github.com/hasrthur/agent-sandbox.nix/commit/96161f709c47ffa4ba0e74b9fd70552dc7005675))
* **test-deep-cwd:** inverse system conditional ([d0ae4d9](https://github.com/hasrthur/agent-sandbox.nix/commit/d0ae4d9bd6969bd603c57bbfe1f234543d999a2a))
* **tests:** use local proxy for tests ([a3cec8f](https://github.com/hasrthur/agent-sandbox.nix/commit/a3cec8f60f61eaa9c4f2014d01d1ff127a528f65))
* use bashInteractive ([2dc4550](https://github.com/hasrthur/agent-sandbox.nix/commit/2dc455059b18efbbdb93c94507e849efad08dc2c))
* use bashInteractive ([6050b01](https://github.com/hasrthur/agent-sandbox.nix/commit/6050b0192e8987a21038fe780ca8443437557ef2))
* wrap bash with --norc and --noprofile ([7884844](https://github.com/hasrthur/agent-sandbox.nix/commit/788484461495bf3aaaf6d99b32478a577b7e708c))
* write test output to repo ([204f895](https://github.com/hasrthur/agent-sandbox.nix/commit/204f895ef65df0d3185599c1cfb9193fa2b66783))

## [5.10.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.9.0...v5.10.0) (2026-11-11)


### Features

* log injected credentials ([#211](https://github.com/archie-judd/agent-sandbox.nix/issues/211)) ([277aeef](https://github.com/archie-judd/agent-sandbox.nix/commit/277aeefb7b566d02cc6c31b5a735d6e610e556ec))

## [5.9.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.8.0...v5.9.0) (2026-11-10)


### Features

* antigravity-cli template ([#209](https://github.com/archie-judd/agent-sandbox.nix/issues/209)) ([20e534a](https://github.com/archie-judd/agent-sandbox.nix/commit/20e534ab2640b1f8fd96fe8a679339b0bd4e19bc))

## [5.8.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.7.1...v5.8.0) (2026-11-10)


### Features

* injectedCredentials ([#207](https://github.com/archie-judd/agent-sandbox.nix/issues/207)) ([dcab096](https://github.com/archie-judd/agent-sandbox.nix/commit/dcab0968a0946a063981e7f6c98b41a102238352))

## [5.7.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.7.0...v5.7.1) (2026-11-05)


### Bug Fixes

* **darwin:** deny process info instead of starting a new session ([#204](https://github.com/archie-judd/agent-sandbox.nix/issues/204)) ([d78cfcd](https://github.com/archie-judd/agent-sandbox.nix/commit/d78cfcd59a99f382bcad9dcf73b530a4751facd0))

## [5.7.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.6.5...v5.7.0) (2026-10-28)


### Features

* **proxy:** allow per-path rules for allowedDomains ([#202](https://github.com/archie-judd/agent-sandbox.nix/issues/202)) ([dbc373b](https://github.com/archie-judd/agent-sandbox.nix/commit/dbc373b49edf2df48626e9360f33e4e76290f4c6))

## [5.6.5](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.6.4...v5.6.5) (2026-10-28)


### Bug Fixes

* **linux:** bind an empty resolv.conf in restricted mode ([#200](https://github.com/archie-judd/agent-sandbox.nix/issues/200)) ([f90f7af](https://github.com/archie-judd/agent-sandbox.nix/commit/f90f7af92792f9d7aeb0fe159c1c15011fc3dbd3))

## [5.6.4](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.6.3...v5.6.4) (2026-10-26)


### Bug Fixes

* **darwin:** let nix 2.34.9 stat the Rosetta runtime ([42ce94b](https://github.com/archie-judd/agent-sandbox.nix/commit/42ce94b657e85dae6a6d5f30d4ee0ee70399ce9f))

## [5.6.3](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.6.2...v5.6.3) (2026-10-18)


### Bug Fixes

* **darwin:** restore resize and Ctrl+Z under the new session ([#194](https://github.com/archie-judd/agent-sandbox.nix/issues/194)) ([407b549](https://github.com/archie-judd/agent-sandbox.nix/commit/407b549e0e1d2440212c5160259a48e26c0af833))

## [5.6.2](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.6.1...v5.6.2) (2026-10-17)


### Bug Fixes

* **darwin:** forward terminal resizes to the sandbox ([#192](https://github.com/archie-judd/agent-sandbox.nix/issues/192)) ([1e09e34](https://github.com/archie-judd/agent-sandbox.nix/commit/1e09e348a5b3304801bae0d97290ff972fe3ab1e))

## [5.6.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.6.0...v5.6.1) (2026-10-14)


### Bug Fixes

* **linux:** disable IPv6 on the sandbox's network interface ([#190](https://github.com/archie-judd/agent-sandbox.nix/issues/190)) ([427b320](https://github.com/archie-judd/agent-sandbox.nix/commit/427b3207c8ddb385b360d1d02a4c6e93fcef0d5c))

## [5.6.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.5.3...v5.6.0) (2026-10-13)


### Features

* print version and session at launch ([#188](https://github.com/archie-judd/agent-sandbox.nix/issues/188)) ([3cb1e72](https://github.com/archie-judd/agent-sandbox.nix/commit/3cb1e72cf9873b9f2ab4f759fea41be1eff4d22b))

## [5.5.3](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.5.2...v5.5.3) (2026-10-12)


### Bug Fixes

* **darwin:** let the sandbox control its new terminal ([8e1e563](https://github.com/archie-judd/agent-sandbox.nix/commit/8e1e563ca1881d2bb6f0c7606a2b205dbba12827))

## [5.5.2](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.5.1...v5.5.2) (2026-10-12)


### Bug Fixes

* **darwin:** start the sandbox in its own session ([#185](https://github.com/archie-judd/agent-sandbox.nix/issues/185)) ([41477e2](https://github.com/archie-judd/agent-sandbox.nix/commit/41477e251b85aca18d3291af73616026a0208435))

## [5.5.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.5.0...v5.5.1) (2026-10-11)


### Bug Fixes

* **darwin:** block procargs reads of host processes ([#183](https://github.com/archie-judd/agent-sandbox.nix/issues/183)) ([9253163](https://github.com/archie-judd/agent-sandbox.nix/commit/925316356e501ccf3a90f427d4df70c302a0fd92))

## [5.5.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.4.1...v5.5.0) (2026-10-06)


### Features

* allowHeadlessBrowsers ([#179](https://github.com/archie-judd/agent-sandbox.nix/issues/179)) ([bfad3a1](https://github.com/archie-judd/agent-sandbox.nix/commit/bfad3a1818d36c62257e0bac3c4b3852b8b7b5c2))

## [5.4.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.4.0...v5.4.1) (2026-09-29)


### Bug Fixes

* launch in a subshell ([#172](https://github.com/archie-judd/agent-sandbox.nix/issues/172)) ([542427e](https://github.com/archie-judd/agent-sandbox.nix/commit/542427ebb50039e00e08f8f928dc55921cac259a))

## [5.4.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.3.1...v5.4.0) (2026-09-28)


### Features

* workspaceDir ([#169](https://github.com/archie-judd/agent-sandbox.nix/issues/169)) ([ee965e5](https://github.com/archie-judd/agent-sandbox.nix/commit/ee965e584f5a94d374ea4e8ea6ee47187ea4e79e))

## [5.3.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.3.0...v5.3.1) (2026-09-26)


### Bug Fixes

* **darwin:** run git hooks when the gitdir is outside the launch directory ([#167](https://github.com/archie-judd/agent-sandbox.nix/issues/167)) ([a70a90c](https://github.com/archie-judd/agent-sandbox.nix/commit/a70a90cb4e730e4edbc49fc6bd76fbaf76b794eb))

## [5.3.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.10...v5.3.0) (2026-09-18)


### Features

* **allowNix:** refuse a trusted user, confirm an unsandboxed daemon ([#163](https://github.com/archie-judd/agent-sandbox.nix/issues/163)) ([0349ba0](https://github.com/archie-judd/agent-sandbox.nix/commit/0349ba0055899b6a39f92ee26c296a1d08603a9a))

## [5.2.10](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.9...v5.2.10) (2026-09-18)


### Bug Fixes

* **darwin:** follow symlink chains ([#160](https://github.com/archie-judd/agent-sandbox.nix/issues/160)) ([4664d0b](https://github.com/archie-judd/agent-sandbox.nix/commit/4664d0b438c37b99628ae1de837ca18bd82fc0a5))

## [5.2.9](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.8...v5.2.9) (2026-09-13)


### Bug Fixes

* correct AUDIT_ARCH_AARCH64 in the AF_UNIX seccomp filter ([21d13be](https://github.com/archie-judd/agent-sandbox.nix/commit/21d13bec3deeb4c427fa0d6c98f97ef4b4e37638))

## [5.2.8](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.7...v5.2.8) (2026-09-13)


### Bug Fixes

* sync shell version ([7097ca1](https://github.com/archie-judd/agent-sandbox.nix/commit/7097ca13cb2129406373abc56816bae38e3b5573))

## [5.2.7](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.6...v5.2.7) (2026-09-13)


### Bug Fixes

* **git:** pin objects/info/alternates in every gitdir ([#153](https://github.com/archie-judd/agent-sandbox.nix/issues/153)) ([2b6647a](https://github.com/archie-judd/agent-sandbox.nix/commit/2b6647a5e2a04675d30833ed7fe269205aacc50b))
* **proxy:** bound upstream dials and cap concurrent connections ([#147](https://github.com/archie-judd/agent-sandbox.nix/issues/147)) ([baa4ae3](https://github.com/archie-judd/agent-sandbox.nix/commit/baa4ae32a3283edcfb5ac8bc0a406a3561c2505a))
* **proxy:** cap request headers, deadline the client handshake, log accept backoff ([#149](https://github.com/archie-judd/agent-sandbox.nix/issues/149)) ([484524a](https://github.com/archie-judd/agent-sandbox.nix/commit/484524a05dab5b9f120ae1e4df3a2d3a958252bf))
* **proxy:** document + test default policy ([#152](https://github.com/archie-judd/agent-sandbox.nix/issues/152)) ([b4e36b2](https://github.com/archie-judd/agent-sandbox.nix/commit/b4e36b2effafc2611f1ea0556db60cee1fa0e7f4))
* **proxy:** fold hostnames ASCII-only and refuse a non-ASCII host ([#151](https://github.com/archie-judd/agent-sandbox.nix/issues/151)) ([019780f](https://github.com/archie-judd/agent-sandbox.nix/commit/019780facaf266b3b7e08b4095ff4857f76ef4d0))
* **proxy:** require the Host header to match the CONNECT host ([#150](https://github.com/archie-judd/agent-sandbox.nix/issues/150)) ([8813839](https://github.com/archie-judd/agent-sandbox.nix/commit/8813839576a062c169589cf68479e9e835293816))

## [5.2.6](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.5...v5.2.6) (2026-09-12)


### Bug Fixes

* **launcher:** always set SANDBOX_PROXY_REDIRECT ([4a30596](https://github.com/archie-judd/agent-sandbox.nix/commit/4a305962be3b27a7807117e12c9905d947aa7eb5))
* **proxy:** disallow , and = in redirects ([#146](https://github.com/archie-judd/agent-sandbox.nix/issues/146)) ([f1397bc](https://github.com/archie-judd/agent-sandbox.nix/commit/f1397bcd42363b046c7cde1273be887d1e26a62c))

## [5.2.5](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.4...v5.2.5) (2026-09-11)


### Bug Fixes

* **linux:** refuse launch for unresolved symlink hop ([#143](https://github.com/archie-judd/agent-sandbox.nix/issues/143)) ([8b72979](https://github.com/archie-judd/agent-sandbox.nix/commit/8b729792aaf573dd8205d095dc5d7e7fa8252ff4))

## [5.2.4](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.3...v5.2.4) (2026-09-11)


### Bug Fixes

* **darwin:** deny writes to read-only paths inside writable ones ([#140](https://github.com/archie-judd/agent-sandbox.nix/issues/140)) ([461ea4b](https://github.com/archie-judd/agent-sandbox.nix/commit/461ea4b0e7a6559d4f380b07bf9b03c8f725388d))

## [5.2.3](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.2...v5.2.3) (2026-09-09)


### Bug Fixes

* set PKG_CONFIG_PATH ([#138](https://github.com/archie-judd/agent-sandbox.nix/issues/138)) ([f801173](https://github.com/archie-judd/agent-sandbox.nix/commit/f8011734d69bbe2fdc0c3f568a7bb3ec98535e47))

## [5.2.2](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.1...v5.2.2) (2026-09-09)


### Bug Fixes

* **allowNix:** ensure daemon ([#135](https://github.com/archie-judd/agent-sandbox.nix/issues/135)) ([8fa9e44](https://github.com/archie-judd/agent-sandbox.nix/commit/8fa9e44e12eefd1bcd63a663c8ed40eec26f6ef1))

## [5.2.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.2.0...v5.2.1) (2026-09-08)


### Bug Fixes

* **darwin:** restrict nix store reads to symlinked files + allowedPackages ([#133](https://github.com/archie-judd/agent-sandbox.nix/issues/133)) ([61bdf31](https://github.com/archie-judd/agent-sandbox.nix/commit/61bdf31ca9426e4f7fb77632e188c172a13fbae9))

## [5.2.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.1.0...v5.2.0) (2026-09-06)


### Features

* **commonTools:** add curl ([e9cd55d](https://github.com/archie-judd/agent-sandbox.nix/commit/e9cd55db3bef751f41de02ea43a16bb3bebbd909))

## [5.1.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v5.0.0...v5.1.0) (2026-09-05)


### Features

* **proxy:** websockets ([#127](https://github.com/archie-judd/agent-sandbox.nix/issues/127)) ([4f1b185](https://github.com/archie-judd/agent-sandbox.nix/commit/4f1b1854f3048c8e0145a8b3c3e847af98214d55))


### Bug Fixes

* **proxy:** write the 101 handshake in a single write ([#129](https://github.com/archie-judd/agent-sandbox.nix/issues/129)) ([e9176ab](https://github.com/archie-judd/agent-sandbox.nix/commit/e9176ab0299d29eb92f6ea5d526088c8bc7f9f90))

## [5.0.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v4.2.0...v5.0.0) (2026-09-05)


### ⚠ BREAKING CHANGES

* rename allowedLocalPorts -> allowedHostPorts

### Features

* allowedInboundPorts — host-to-sandbox TCP port forwards ([#116](https://github.com/archie-judd/agent-sandbox.nix/issues/116)) ([c84478a](https://github.com/archie-judd/agent-sandbox.nix/commit/c84478ae3ccd0f6fbbc790ef2ce6907e9e4390db))
* rename allowedLocalPorts -&gt; allowedHostPorts ([99b75c8](https://github.com/archie-judd/agent-sandbox.nix/commit/99b75c820c9a2b6055250f4de6eb78a7641f855b))

## [4.2.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v4.1.4...v4.2.0) (2026-09-05)


### Features

* **proxy:** log each allowed host on first contact ([#123](https://github.com/archie-judd/agent-sandbox.nix/issues/123)) ([a69b1ee](https://github.com/archie-judd/agent-sandbox.nix/commit/a69b1eed7e15cede8ba52f90e80df2c6e355ff7c))


### Bug Fixes

* **darwin:** add the session tmpdir to the AF_UNIX socket scope ([#121](https://github.com/archie-judd/agent-sandbox.nix/issues/121)) ([b8c1e45](https://github.com/archie-judd/agent-sandbox.nix/commit/b8c1e459dc9b4765f403e4e1437797a335f0bc1f))

## [4.1.4](https://github.com/archie-judd/agent-sandbox.nix/compare/v4.1.3...v4.1.4) (2026-09-02)


### Bug Fixes

* **darwin:** set CLAUDE_CODE_TMPDIR ([#118](https://github.com/archie-judd/agent-sandbox.nix/issues/118)) ([6e517f5](https://github.com/archie-judd/agent-sandbox.nix/commit/6e517f5e2181bcf227675642456560365cccf47d))

## [4.1.3](https://github.com/archie-judd/agent-sandbox.nix/compare/v4.1.2...v4.1.3) (2026-08-28)


### Bug Fixes

* **git:** grant the work tree root rather than the common git dir's parent ([#102](https://github.com/archie-judd/agent-sandbox.nix/issues/102)) ([bd7a87a](https://github.com/archie-judd/agent-sandbox.nix/commit/bd7a87a5189011adb64ab694ae1014c99ad84b03))

## [4.1.2](https://github.com/archie-judd/agent-sandbox.nix/compare/v4.1.1...v4.1.2) (2026-08-28)


### Bug Fixes

* **darwin:** write $HOME and $TMPDIR to session state dir ([#103](https://github.com/archie-judd/agent-sandbox.nix/issues/103)) ([0f7647a](https://github.com/archie-judd/agent-sandbox.nix/commit/0f7647aacf3bf4a9a45b227edcd9192f426590d9))

## [4.1.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v4.1.0...v4.1.1) (2026-08-28)


### Bug Fixes

* **launch log:** make bwrap.args readable ([ffd47e0](https://github.com/archie-judd/agent-sandbox.nix/commit/ffd47e0db1b4b4443fef8727f02213932f1bdb86))

## [4.1.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v4.0.0...v4.1.0) (2026-08-27)


### Features

* startup log ([#97](https://github.com/archie-judd/agent-sandbox.nix/issues/97)) ([9451d86](https://github.com/archie-judd/agent-sandbox.nix/commit/9451d86e545490c162bc9b2bbea2254a06f4d0c4))
* version ([#99](https://github.com/archie-judd/agent-sandbox.nix/issues/99)) ([b2217b6](https://github.com/archie-judd/agent-sandbox.nix/commit/b2217b6f91e050608cc5a9f1c75eab54f5859103))

## [4.0.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v3.0.0...v4.0.0) (2026-08-27)


### ⚠ BREAKING CHANGES

* Linux sandboxes that implicitly relied on AF_UNIX sockets must set allowUnixSockets = true.

### darwin

* allow AF_UNIX bind/connect scoped to writable dirs (CWD + rwD… ([#92](https://github.com/archie-judd/agent-sandbox.nix/issues/92)) ([dbca506](https://github.com/archie-judd/agent-sandbox.nix/commit/dbca50647f4e9f5933b3c573ee7b6e4552526129))

## [3.0.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.5.1...v3.0.0) (2026-08-25)


### ⚠ BREAKING CHANGES

* migrate to python

### Features

* migrate to python ([7695a5f](https://github.com/archie-judd/agent-sandbox.nix/commit/7695a5f6319275aa75a816787424f57fdd83bfae))

## [2.5.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.5.0...v2.5.1) (2026-08-23)


### Bug Fixes

* **darwin:** refuse binds nested inside another declared bind ([#87](https://github.com/archie-judd/agent-sandbox.nix/issues/87)) ([bd51a8b](https://github.com/archie-judd/agent-sandbox.nix/commit/bd51a8b030c796cbbdbab89c76fab79fe7c26e57))

## [2.5.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.4.1...v2.5.0) (2026-08-23)


### Features

* TODO.md ([8543e1e](https://github.com/archie-judd/agent-sandbox.nix/commit/8543e1eb37e446259cecc473f4f4ea9d8d0d4062))


### Bug Fixes

* **proxy:** disallow addresses resolving to loopback ([#86](https://github.com/archie-judd/agent-sandbox.nix/issues/86)) ([8265aa5](https://github.com/archie-judd/agent-sandbox.nix/commit/8265aa50592d5c3aa3559e7115702bcb8ac399a0))

## [2.4.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.4.0...v2.4.1) (2026-08-20)


### Bug Fixes

* protect the whole repo's gitdir from hook injection ([#79](https://github.com/archie-judd/agent-sandbox.nix/issues/79)) ([147d9af](https://github.com/archie-judd/agent-sandbox.nix/commit/147d9af8c8936c9937767f698d2774a5709ddbe7))

## [2.4.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.3.0...v2.4.0) (2026-08-20)


### Features

* allow launching from home ([#77](https://github.com/archie-judd/agent-sandbox.nix/issues/77)) ([d147b38](https://github.com/archie-judd/agent-sandbox.nix/commit/d147b38fe62a74f70036aa9203df3b7e8956375d))

## [2.3.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.2.1...v2.3.0) (2026-08-19)


### Features

* relicense under MIT ([1462f69](https://github.com/archie-judd/agent-sandbox.nix/commit/1462f69154f33efd6b809ed1a598dd405673128c))

## [2.2.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.2.0...v2.2.1) (2026-07-13)


### Bug Fixes

* **Darwin:** fix whitespace string issue for allowedLocalPort seatbelt str ([637511e](https://github.com/archie-judd/agent-sandbox.nix/commit/637511e5fc7613804e6a4643eea9c73494e35c53))

## [2.2.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.1.0...v2.2.0) (2026-07-10)


### Features

* **allowLocalPorts:** Add cross-platform localNetworkAccess targets ([#68](https://github.com/archie-judd/agent-sandbox.nix/issues/68)) ([3a0930a](https://github.com/archie-judd/agent-sandbox.nix/commit/3a0930a2dc9e24d9f93d4c7998f36b1a0a5d272d))

## [2.1.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.0.1...v2.1.0) (2026-06-18)


### Features

* add claude-nix shell.nix ([76dae03](https://github.com/archie-judd/agent-sandbox.nix/commit/76dae037721b1e9fc83f483631fa84af60acc34d))
* **darwin:** allowNix ([6cea43e](https://github.com/archie-judd/agent-sandbox.nix/commit/6cea43e701186fd8d636bf39f871f06a1f5af0bc))
* **linux:** allowNix ([9f0219d](https://github.com/archie-judd/agent-sandbox.nix/commit/9f0219da1156a206573dcf5ddd117bf10c43aee5))
* **README:** allowNix ([bcea1b2](https://github.com/archie-judd/agent-sandbox.nix/commit/bcea1b28f4ed2c4aff2d2672152d6963f0e5604d))


### Bug Fixes

* **darwin:** resolve real nix daemon socket path ([2e9dd51](https://github.com/archie-judd/agent-sandbox.nix/commit/2e9dd5171b70bb0c727b16f21bdf2351fd285f6c))

## [2.0.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v2.0.0...v2.0.1) (2026-06-16)


### Bug Fixes

* **linux:** bind roFile/rwFile symlinks at their declared paths ([4c9cac0](https://github.com/archie-judd/agent-sandbox.nix/commit/4c9cac00437dbfcdf26cf013e434da53a7954fa2))
* **linux:** don't follow symlinks when binding files ([67e7018](https://github.com/archie-judd/agent-sandbox.nix/commit/67e70185e91eb98983a2dbeea73d870d57ef5477))

## [2.0.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v1.0.0...v2.0.0) (2026-06-13)


### ⚠ BREAKING CHANGES

* declared rwDirs/rwFiles must exist before launch
* fail closed on git identity instead of fabricating one

### Features

* declared rwDirs/rwFiles must exist before launch ([1342f80](https://github.com/archie-judd/agent-sandbox.nix/commit/1342f808651dd3fe71b28a033158dd76b1df0117))
* fail closed on git identity instead of fabricating one ([f1122d1](https://github.com/archie-judd/agent-sandbox.nix/commit/f1122d1b920831ee6463420eefd8b8206f46a14e))
* roDirs and roFiles read-only bind primitives ([54066d0](https://github.com/archie-judd/agent-sandbox.nix/commit/54066d013545451f1bda2974497ef000c4ae2608))
* roDirs and roFiles read-only bind primitives ([5f8ce44](https://github.com/archie-judd/agent-sandbox.nix/commit/5f8ce44c26c0cfca299f9ac9c534dfe9d1be4773))


### Bug Fixes

* resolve resolv.conf on ubuntu ([cc2d145](https://github.com/archie-judd/agent-sandbox.nix/commit/cc2d1453b43269dc8c3869e4cee160cb7a1d385c))

## [1.0.0](https://github.com/archie-judd/agent-sandbox.nix/compare/v0.1.1...v1.0.0) (2026-06-12)


### ⚠ BREAKING CHANGES

* Renamed extraEnv → env. Pure rename; semantics unchanged.

### Features

* rename API args and replace restrictNetwork with allowedDomains ([a2ee921](https://github.com/archie-judd/agent-sandbox.nix/commit/a2ee921d1ff2b158d8391fb5f22ca5774d5955f8))

## [0.1.1](https://github.com/archie-judd/agent-sandbox.nix/compare/v0.1.0...v0.1.1) (2026-06-10)


### Bug Fixes

* disable .git discovery when $HOME==$REPO_ROOT ([72ac65c](https://github.com/archie-judd/agent-sandbox.nix/commit/72ac65c108761af326e8403c40a736ae755a6b92))
