type formTypes =
  | ContactName
  | ContactType
  | ContactUrl;

type formState = {
  mutable contactName: Form.t,
  mutable contactType: Form.t,
  mutable contactUrl: Form.t,
};

let getInitialFormState = () => {
  {contactName: Str(""), contactType: Str(""), contactUrl: Str("")};
};

module FormConfig = {
  type state = formState;
  type field = formTypes;
  let update = (field, value, state) => {
    switch (field, value) {
    | (ContactName, v) => {...state, contactName: v}
    | (ContactType, v) => {...state, contactType: v}
    | (ContactUrl, v) => {...state, contactUrl: v}
    };
  };
  let get = (field, state) => {
    switch (field) {
    | ContactName => state.contactName
    | ContactType => state.contactType
    | ContactUrl => state.contactUrl
    };
  };
};

let emptyMsg = "Field is required";

let rules = [
  (ContactName, [(Form.NotEmpty, emptyMsg)]),
  (ContactType, [(Form.NotEmpty, emptyMsg)]),
  (ContactUrl, [(Form.NotEmpty, emptyMsg)]),
];

module ContactForm = Form.FormComponent(FormConfig);

let payloadOfState = (~inputContact=?, values) => {
  let payload = Js.Dict.empty();
  switch (inputContact) {
  | Some(contact) =>
    Models.Contact.(
      switch (contact) {
      | Some(contact_) =>
        FormHelpers.setJsonKey(
          payload,
          "contact_id",
          Str(contact_.contactId),
        )
      | None => ()
      }
    )

  | None => ()
  };

  FormHelpers.setJsonKey(payload, "contact_name", values.contactName);
  FormHelpers.setJsonKey(payload, "contact_type", values.contactType);
  FormHelpers.setJsonKey(payload, "contact_url", values.contactUrl);

  payload;
};

[@react.component]
let make = (~submitContact, ~inputContact: option(Models.Contact.t)=?) => {
  let (submitted, setSubmitted) = React.useState(() => false);
  let (errorMsg, setErrorMsg) = React.useState(() => "");
  let (tryTestMsg, setTryTestMsg) = React.useState(() => "");

  let postCallback = resp => {
    switch (resp) {
    | Api.Error(msg) => setErrorMsg(_ => msg)
    | Api.Success(_) => Paths.goToContacts()
    | Api.SuccessJSON(_) => Paths.goToContacts()
    };
  };

  let testContactCallback = resp => {
    switch (resp) {
    | Api.Error(msg) => setErrorMsg(_ => msg)
    | Api.Success(msg) => setTryTestMsg(_ => msg)
    | Api.SuccessJSON(_) => Paths.goToContacts()
    };
  };

  let handleSubmit = (e, values, errors) => {
    ReactEvent.Form.preventDefault(e);
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      let payload = payloadOfState(values, ~inputContact);
      submitContact(payload, postCallback);
    };
  };

  let handleTryContact = (values, errors) => {
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      setTryTestMsg(_ => "Loading...");
      let payload = payloadOfState(values);
      Api.tryContact(payload, testContactCallback);
    };
  };

  <div className="px-6 py-2 lg:w-1/2">
    <ContactForm
      initialState={
                     let init = getInitialFormState();
                     switch (inputContact) {
                     | None => ()
                     | Some(contact) =>
                       init.contactName = Str(contact.contactName);
                       init.contactType = Str(contact.contactType);
                       init.contactUrl = Str(contact.contactUrl);
                     };
                     init;
                   }
      rules
      render={(f: ContactForm.form) =>
        <form
          className="w-full"
          onSubmit={e => handleSubmit(e, f.form.values, f.form.errors)}>
          <div className="flex flex-wrap -mx-3 mb-3">
            <div className="w-full md:w-1/2 px-3 mb-6">
              <label
                className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                {"Type" |> React.string}
              </label>
              <div className="relative">
                <select
                  value={
                    switch (f.form.values.contactType) {
                    | Str(s) => s
                    | _ => ""
                    }
                  }
                  onChange={e =>
                    Form.Str(ReactEvent.Form.target(e)##value)
                    |> f.handleChange(ContactType)
                  }
                  className="block appearance-none w-full bg-gray-200 border border-gray-200 text-gray-700 py-3 px-4 pr-8 rounded leading-tight focus:outline-none focus:bg-white focus:border-gray-500">
                  <option value="" disabled=true hidden=true>
                    {"Choose contact type" |> React.string}
                  </option>
                  <option value="smtp"> {"smtp" |> React.string} </option>
                  <option value="http"> {"http" |> React.string} </option>
                </select>
                <div
                  className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-700">
                  <svg
                    className="fill-current h-4 w-4"
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 20 20">
                    <path
                      d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z"
                    />
                  </svg>
                </div>
              </div>
              {let errorMsg =
                 FormHelpers.getError(ContactType, f.form.errors);
               errorMsg != React.null && submitted
                 ? <p className="text-red-500 text-xs italic"> errorMsg </p>
                 : <p className="text-gray-600 text-xs italic">
                     {"Contact's type (*)" |> React.string}
                   </p>}
            </div>
            <div className="w-full md:w-1/2 px-3">
              <label
                className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                {"Name" |> React.string}
              </label>
              <input
                value={
                  switch (f.form.values.contactName) {
                  | Str(s) => s
                  | _ => ""
                  }
                }
                onChange={e =>
                  Form.Str(ReactEvent.Form.target(e)##value)
                  |> f.handleChange(ContactName)
                }
                className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
                type_="text"
                placeholder="The name of the contact"
              />
              {FormHelpers.getError(ContactName, f.form.errors) != React.null
               && submitted
                 ? <p className="text-red-500 text-xs italic">
                     {FormHelpers.getError(ContactName, f.form.errors)}
                   </p>
                 : <p className="text-gray-600 text-xs italic">
                     {"Contact's name, no functional meaning (*)"
                      |> React.string}
                   </p>}
            </div>
            <div className="w-full px-3">
              <label
                className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                {"url" |> React.string}
              </label>
              <input
                value={
                  switch (f.form.values.contactUrl) {
                  | Str(s) => s
                  | _ => ""
                  }
                }
                onChange={e =>
                  Form.Str(ReactEvent.Form.target(e)##value)
                  |> f.handleChange(ContactUrl)
                }
                className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500"
                type_="text"
                placeholder={
                  switch (f.form.values.contactType) {
                  | Str(type_) =>
                    switch (type_) {
                    | "smtp" => "example@gmail.com"
                    | _ => "https://example.com"
                    }
                  | _ => "https://example.com"
                  }
                }
              />
              {FormHelpers.getError(ContactUrl, f.form.errors) != React.null
               && submitted
                 ? <p className="text-red-500 text-xs italic">
                     {FormHelpers.getError(ContactUrl, f.form.errors)}
                   </p>
                 : <p className="text-gray-600 text-xs italic">
                     {"Email/Hook url (*)" |> React.string}
                   </p>}
            </div>
          </div>
          <button
            type_="button"
            onClick={_ => handleTryContact(f.form.values, f.form.errors)}
            className="mr-1 bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
            {"Test contact" |> React.string}
          </button>
          <button
            type_="submit"
            className="m-1 bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded">
            {"Submit" |> React.string}
          </button>
          {tryTestMsg != ""
             ? <p className="text-gray-600 m-1">
                 {tryTestMsg |> React.string}
               </p>
             : React.null}
          {errorMsg != ""
             ? <p className="text-red-500 text-xs italic m-1">
                 {"Error posting contact: " ++ errorMsg |> React.string}
               </p>
             : React.null}
        </form>
      }
    />
  </div>;
};