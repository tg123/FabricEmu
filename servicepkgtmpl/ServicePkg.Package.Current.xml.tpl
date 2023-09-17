<?xml version="1.0" encoding="utf-8"?>
<ServicePackage xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" Name="ServicePkg" ManifestVersion="1.0.0" RolloutVersion="1.0" xmlns="http://schemas.microsoft.com/2011/01/fabric">
	<DigestedServiceTypes RolloutVersion="1.0">
		<ServiceTypes>
		</ServiceTypes>
	</DigestedServiceTypes>
	<DigestedCodePackage RolloutVersion="1.0">
		<CodePackage Name="Code" Version="1.0.0">
			<EntryPoint>
			<!-- do not understand why <EntryPoint/> crash, wierd sf code is -->
			</EntryPoint>
		</CodePackage>
	</DigestedCodePackage>
	<DigestedConfigPackage RolloutVersion="1.0">
		<ConfigPackage Name="Config" Version="1.0" />
	</DigestedConfigPackage>
	<DigestedResources RolloutVersion="1.0" />
	<Diagnostics />
</ServicePackage>