<?xml version="1.0" encoding="utf-8"?>
<ServiceManifest xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" Name="" Version="1.0.0" xmlns="http://schemas.microsoft.com/2011/01/fabric">
  <ServiceTypes>
    {{range .ServiceTypes.StatefulServiceType}}
    <StatefulServiceType ServiceTypeName="{{.ServiceTypeName}}" />
    {{end}}
  </ServiceTypes>
</ServiceManifest>