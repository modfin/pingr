type sshFormTypes =
  /* Poll tests */
  | TestName
  | Interval
  | Timeout
  | Url
  /* Contacts */
  | Contacts
  /* SSH */
  | Username
  | Port
  | CredentialType
  | Password
  | Key;

type sshFormState = {
  /* Poll tests */
  mutable testName: Form.t,
  mutable interval: Form.t,
  mutable timeout: Form.t,
  mutable url: Form.t,
  /* Contacts */
  mutable contacts: Form.t,
  /* SSH */
  mutable port: Form.t,
  mutable credentialType: Form.t,
  mutable username: Form.t,
  mutable password: Form.t,
  mutable key: Form.t,
};

let getInitialFormState = () => {
  testName: Str(""),
  interval: Int(0),
  timeout: Int(0),
  url: Str(""),
  contacts: TupleList([]),
  port: Str(""),
  credentialType: Str(""),
  username: Str(""),
  password: Str(""),
  key: Str(""),
};

module SSHFormConfig = {
  type field = sshFormTypes;
  type state = sshFormState;
  let update = (field, value, state) => {
    switch (field, value) {
    | (TestName, v) => {...state, testName: v}
    | (Interval, v) => {...state, interval: v}
    | (Timeout, v) => {...state, timeout: v}
    | (Url, v) => {...state, url: v}
    | (Contacts, v) => {...state, contacts: v}
    | (Port, v) => {...state, port: v}
    | (CredentialType, v) => {...state, credentialType: v}
    | (Username, v) => {...state, username: v}
    | (Password, v) => {...state, password: v}
    | (Key, v) => {...state, key: v}
    };
  };
  let get = (field, state) => {
    switch (field) {
    | TestName => state.testName
    | Interval => state.interval
    | Timeout => state.timeout
    | Url => state.url
    | Contacts => state.contacts
    | Port => state.port
    | CredentialType => state.credentialType
    | Username => state.username
    | Password => state.password
    | Key => state.key
    };
  };
};

module SSHForm = Form.FormComponent(SSHFormConfig);

let pswValidation = (value, values) => {
  switch (values.credentialType) {
  | Str("userpass") => value != Form.Str("")
  | _ => true
  };
};

let keyValidation = (value, values) => {
  switch (values.credentialType) {
  | Str("key") => value != Form.Str("")
  | _ => true
  };
};

let defaultRules = [
  (TestName, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (Interval, [(Form.NotEmpty, FormHelpers.aboveZero)]),
  (Timeout, [(Form.NotEmpty, FormHelpers.aboveZero)]),
  (Url, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (
    Contacts,
    [
      (
        Form.Custom(FormHelpers.validContactThreshold),
        FormHelpers.thresholdMsg,
      ),
    ],
  ),
  (Port, [(Form.NotEmpty, FormHelpers.aboveZero)]),
  (CredentialType, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (Username, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
];

let getTestPayload = (~inputTest=?, values) => {
  let payload = Js.Dict.empty();
  switch (inputTest) {
  | Some(test) =>
    Models.Test.(
      switch (test) {
      | Some(test_) =>
        FormHelpers.setJsonKey(payload, "test_id", Str(test_.testId));
        Js.Dict.set(payload, "active", Js.Json.boolean(test_.active));
      | None => Js.Dict.set(payload, "active", Js.Json.boolean(true))
      }
    )

  | None => ()
  };

  FormHelpers.setJsonKey(payload, "test_type", Str("SSH"));
  FormHelpers.setJsonKey(payload, "test_name", values.testName);
  FormHelpers.setJsonKey(payload, "timeout", values.timeout);
  FormHelpers.setJsonKey(payload, "url", values.url);
  FormHelpers.setJsonKey(payload, "interval", values.interval);

  let blob = Js.Dict.empty();
  FormHelpers.setJsonKey(blob, "username", values.username);
  FormHelpers.setJsonKey(blob, "port", values.port);
  FormHelpers.setJsonKey(blob, "credential_type", values.credentialType);
  FormHelpers.setJsonKey(
    blob,
    "credential",
    switch (values.credentialType) {
    | Str("key") => values.key
    | Str("userpass") => values.password
    | _ => Str(".")
    },
  );
  Js.Dict.set(payload, "blob", Js.Json.object_(blob));

  payload;
};

[@react.component]
let make =
    (
      ~submitTest,
      ~submitContacts,
      ~inputTest: option(Models.Test.t)=?,
      ~inputTestContacts=?,
    ) => {
  let (submitted, setSubmitted) = React.useState(_ => false);
  let (submitError, setSubmitError) = React.useState(() => "");
  let (tryTestMsg, setTryTestMsg) = React.useState(() => "");

  let rules =
    switch (inputTest) {
    | None =>
      defaultRules
      @ [
        (Password, [(Form.Custom(pswValidation), FormHelpers.emptyMsg)]),
        (Key, [(Form.Custom(keyValidation), FormHelpers.emptyMsg)]),
      ]
    | Some(test) =>
      switch (test.specific) {
      | SSH(ssh) =>
        switch (ssh.credentialType) {
        | "userpass" =>
          defaultRules
          @ [(Key, [(Form.Custom(keyValidation), FormHelpers.emptyMsg)])]
        | "key" =>
          defaultRules
          @ [
            (
              Password,
              [(Form.Custom(pswValidation), FormHelpers.emptyMsg)],
            ),
          ]
        | _ => defaultRules
        }
      | _ => defaultRules
      }
    };

  let rec postCallback = (values, resp) => {
    switch (resp) {
    | Api.Error(msg) => setSubmitError(_ => msg)
    | Api.SuccessJSON(testJson) =>
      let testContacts =
        switch (values.contacts) {
        | TupleList(l) => l
        | _ => []
        };
      if (List.length(testContacts) > 0) {
        let id = Models.Decode.test(testJson).testId;
        submitContacts(
          FormHelpers.getContactsPayload(id, TupleList(testContacts)),
          postCallback(values),
        );
      } else {
        Paths.goToTests();
      };
    | Success(_) => Paths.goToTests()
    };
  };

  let tryTestCallback = resp => {
    switch (resp) {
    | Api.Error(msg) =>
      setSubmitError(_ => msg);
      setTryTestMsg(_ => "");
    | Api.Success(msg) =>
      setTryTestMsg(_ => msg);
      setSubmitError(_ => "");
    | Api.SuccessJSON(_json) => ()
    };
  };

  let handleSubmit = (e, values, errors) => {
    ReactEvent.Form.preventDefault(e);
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      let payload = getTestPayload(values, ~inputTest);
      submitTest(payload, postCallback(values));
    };
  };

  let handleTryTest = (values, errors) => {
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      setTryTestMsg(_ => "Loading...");
      let payload = getTestPayload(values);
      Api.tryTest(payload, tryTestCallback);
    };
  };
  <SSHForm
    rules
    initialState={
                   let init = getInitialFormState();
                   switch (inputTest) {
                   | None => ()
                   | Some(test) =>
                     init.testName = Str(test.testName);
                     init.interval = Int(test.interval);
                     init.timeout = Int(test.timeout);
                     init.url = Str(test.url);
                     switch (test.specific) {
                     | SSH(ssh) =>
                       init.port = Str(ssh.port);
                       init.credentialType = Str(ssh.credentialType);
                       init.username = Str(ssh.username);
                     | _ => ()
                     };
                   };
                   switch (inputTestContacts) {
                   | None => ()
                   | Some(testContacts) =>
                     init.contacts =
                       TupleList(
                         testContacts
                         |> List.map(testContact => {
                              Models.TestContact.(
                                testContact.contactId,
                                string_of_int(testContact.threshold),
                              )
                            }),
                       )
                   };
                   init;
                 }
    render={(f: SSHForm.form) =>
      <form
        onSubmit={e => handleSubmit(e, f.form.values, f.form.errors)}
        className="w-full">
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Text
            width=Full
            label="Name"
            placeholder="Some name"
            infoText="Test's name, no functional meaning (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(TestName, f.form.errors) : React.null
            }
            value={f.form.values.testName}
            onChange={v => v |> f.handleChange(TestName)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Number
            width=Half
            label="Interval (s)"
            infoText="The number of seconds between each test (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Interval, f.form.errors) : React.null
            }
            value={f.form.values.interval}
            onChange={v => v |> f.handleChange(Interval)}
          />
          <FormInput
            type_=Number
            width=Half
            label="Timeout (s)"
            infoText="The number of seconds before test times out (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Timeout, f.form.errors) : React.null
            }
            value={f.form.values.timeout}
            onChange={v => v |> f.handleChange(Timeout)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Text
            width=Full
            label="Hostname"
            infoText="Hostname of SSH server (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Url, f.form.errors) : React.null
            }
            placeholder="google.com"
            value={f.form.values.url}
            onChange={v => v |> f.handleChange(Url)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Text
            width=Full
            label="Port"
            infoText="Port of SSH server (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Port, f.form.errors) : React.null
            }
            placeholder="22"
            value={f.form.values.port}
            onChange={v => v |> f.handleChange(Port)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Text
            width=Full
            label="Username"
            infoText="Username used for authentication (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Username, f.form.errors) : React.null
            }
            value={f.form.values.username}
            onChange={v => v |> f.handleChange(Username)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormSelect
            width=Full
            label="Credential type"
            placeholder="Choose authentication method"
            infoText="Method which authentication will be made (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(CredentialType, f.form.errors)
                : React.null
            }
            options=[
              ("Username and password", "userpass"),
              ("Private key", "key"),
            ]
            onChange={v => v |> f.handleChange(CredentialType)}
            value={f.form.values.credentialType}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          {switch (f.form.values.credentialType) {
           | Str("key") =>
             <FormTextarea
               label="Key"
               placeholder={
                 switch (inputTest) {
                 | None => "-----BEGIN OPENSSH PRIVATE KEY-----......."
                 | Some(test) =>
                   switch (test.specific) {
                   | SSH(ssh) =>
                     ssh.credentialType == "key"
                       ? "Leave blank not to update"
                       : "-----BEGIN OPENSSH PRIVATE KEY-----......."
                   | _ => "-----BEGIN OPENSSH PRIVATE KEY-----......."
                   }
                 }
               }
               value={f.form.values.key}
               onChange={v => v |> f.handleChange(Key)}
               errorMsg={
                 submitted
                   ? FormHelpers.getError(Key, f.form.errors) : React.null
               }
               infoText="Key which will be used to autheticate (*)"
             />
           | Str("userpass") =>
             <FormInput
               type_=Password
               width=Full
               label="Password"
               infoText="Password used for authentication (*)"
               placeholder={
                 switch (inputTest) {
                 | None => ""
                 | Some(test) =>
                   switch (test.specific) {
                   | SSH(ssh) =>
                     ssh.credentialType == "userpass"
                       ? "Leave blank not to update" : ""
                   | _ => ""
                   }
                 }
               }
               errorMsg={
                 submitted
                   ? FormHelpers.getError(Password, f.form.errors)
                   : React.null
               }
               value={f.form.values.password}
               onChange={v => v |> f.handleChange(Password)}
             />
           | _ => React.null
           }}
        </div>
        <FormTestContacts
          errorMsg={
            submitted
              ? FormHelpers.getError(Contacts, f.form.errors) : React.null
          }
          value={
            switch (f.form.values.contacts) {
            | TupleList(l) => l
            | _ => []
            }
          }
          onChange={v => v |> f.handleChange(Contacts)}
        />
        <button
          type_="button"
          onClick={_ => handleTryTest(f.form.values, f.form.errors)}
          className="mr-1 bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
          {"Try test" |> React.string}
        </button>
        <button
          type_="submit"
          className="m-1 bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded">
          {"Submit" |> React.string}
        </button>
        {tryTestMsg != ""
           ? <p className="text-gray-600 mb-2">
               {tryTestMsg |> React.string}
             </p>
           : React.null}
        {submitError != ""
           ? <p className="text-red-500">
               {"Error posting test: " ++ submitError |> React.string}
             </p>
           : React.null}
        <div className="h-32" />
      </form>
    }
  />;
};