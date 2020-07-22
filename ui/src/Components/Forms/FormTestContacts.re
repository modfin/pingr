module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(list(Models.Contact.t))
  | LoadFail(string);

[@react.component]
let make = () => {
  let (testContacts: list((string, string)), setTestContacts) =
    React.useState(_ => []);

  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
        | LoadData => Loadable.Loading
        | LoadSuccess(contacts) => Loadable.Success(contacts)
        | LoadFail(msg) => Loadable.Failed(msg)
        },
      Loadable.Loading,
    );

  React.useEffect0(() => {
    let cb = result => {
      switch (result) {
      | None => dispatch(LoadFail("None"))
      | Some(contacts) => dispatch(LoadSuccess(contacts))
      };
    };
    Api.fetchContactsWithCallback(cb);
    None;
  });

  let updateTestContacts = (checked, contact: (string, string)) =>
    if (checked) {
      setTestContacts(prev => prev @ [contact]);
    } else {
      setTestContacts(prev =>
        prev |> List.filter(testContact => testContact != contact)
      );
    };

  let updateThreshold = (contact: (string, string)) => {
    let id = fst(contact);
    setTestContacts(prev =>
      prev
      |> List.map(((id2, threshold)) =>
           if (id == id2) {
             contact;
           } else {
             (id2, threshold);
           }
         )
    );
  };

  let isChecked = id =>
    switch (testContacts |> List.find(contact => fst(contact) == id)) {
    | exception Not_found => false
    | _contact => true
    };
  let getThreshold = id =>
    switch (testContacts |> List.find(contact => fst(contact) == id)) {
    | exception Not_found => ""
    | testContact => fst(testContact)
    };

  let validThreshold = () => {
    switch (
      testContacts
      |> List.find(testContact => int_of_string(snd(testContact)) <= 0)
    ) {
    | exception Not_found => true
    | _testContact => false
    };
  };

  <div className="w-full">
    <label
      className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
      {"contacts" |> React.string}
    </label>
    {switch (state) {
     | Loading => <div> {ReasonReact.string("Loading...")} </div>
     | Failed(_) => "No contacts. Go to contacts and add some" |> React.string
     | Success(contacts) =>
       <table className="table-auto text-left bg-gray-200 rounded">
         <thead>
           <tr>
             <th className="px-4 py-2 border"> {"" |> React.string} </th>
             <th className="px-4 py-2 border"> {"Name" |> React.string} </th>
             <th className="px-4 py-2 border">
               {"Threshold" |> React.string}
             </th>
           </tr>
         </thead>
         <tbody>
           Models.Contact.(
             {contacts
              |> List.map(contact =>
                   <tr key={contact.contactId}>
                     <td className="border px-4 py-2">
                       <input
                         checked={isChecked(contact.contactId)}
                         onChange={e =>
                           updateTestContacts(
                             ReactEvent.Form.target(e)##checked,
                             (contact.contactId, "0"),
                           )
                         }
                         className="mr-2 leading-tight"
                         type_="checkbox"
                       />
                     </td>
                     <td className="border px-4 py-2">
                       {contact.contactName |> React.string}
                     </td>
                     <td
                       className="block w-full text-gray-700 border py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500">
                       <input
                         type_="number"
                         min="0"
                         onChange={e =>
                           updateThreshold((
                             contact.contactId,
                             ReactEvent.Form.target(e)##value,
                           ))
                         }
                         value={getThreshold(contact.contactId)}
                         disabled={!isChecked(contact.contactId)}
                       />
                     </td>
                   </tr>
                 )
              |> Array.of_list
              |> React.array}
           )
         </tbody>
       </table>
     }}
    {!validThreshold()
       ? <p className="text-red-500 text-xs italic">
           {"Threshold has to be > 0" |> React.string}
         </p>
       : <p className="text-gray-600 text-xs italic">
           {"Who should be contacted upon error and after how many consecutive test failures"
            |> React.string}
         </p>}
  </div>;
};