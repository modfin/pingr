let testTypes = [
  ("Poll - DNS", "DNS"),
  ("Poll - HTTP", "HTTP"),
  ("Poll - Ping", "Ping"),
  ("Poll - Prometheus", "Prometheus"),
  ("Poll - SSH", "SSH"),
  ("Poll - TCP", "TCP"),
  ("Poll - TLS", "TLS"),
  ("Push - HTTP", "HTTPPush"),
  ("Push - Prometheus", "PrometheusPush"),
];

[@react.component]
let make = () => {
  let (testType, setTestType) = React.useState(_ => Form.Str(""));
  <>
    <div className="relative bg-gray-400 my-4 p-1">
      <p className="text-xl font-bold mx-4"> {"New test" |> React.string} </p>
    </div>
    <div className="px-4 pt-4 lg:w-1/2">
      <div className="flex flex-wrap -mx-3 mb-6">
        <FormSelect
          width=Full
          label="type"
          placeholder="Choose test type"
          infoText="Test's type (*)"
          errorMsg=React.null
          options=testTypes
          onChange={v => setTestType(_ => v)}
          value=testType
        />
      </div>
      {switch (testType) {
       | Str("HTTP") =>
         <HTTPForm
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | Str("Ping") =>
         <PingForm
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | Str("TCP") =>
         <PortTestForm
           testType="TCP"
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | Str("TLS") =>
         <PortTestForm
           testType="TLS"
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | Str("Prometheus") =>
         <PrometheusForm
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | Str("DNS") =>
         <DNSForm
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | Str("SSH") =>
         <SSHForm
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | Str("HTTPPush") =>
         <HTTPPushForm
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | Str("PrometheusPush") =>
         <PrometheusPushForm
           submitTest=Api.postTest
           submitContacts=Api.postTestContacts
         />
       | _ => React.null
       }}
    </div>
  </>;
};