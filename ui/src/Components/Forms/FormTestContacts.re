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
let make = (~value, ~onChange, ~errorMsg) => {
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

  let updateTestContacts = (checked, contact: (string, string), currContacts) =>
    if (checked) {
      currContacts @ [contact];
    } else {
      currContacts |> List.filter(testContact => testContact != contact);
    };

  let updateThreshold = (contact: (string, string), currContacts) => {
    let id = fst(contact);
    currContacts
    |> List.map(((id2, threshold)) =>
         if (id == id2) {
           contact;
         } else {
           (id2, threshold);
         }
       );
  };

  let isChecked = (id, currContacts) =>
    switch (currContacts |> List.find(contact => fst(contact) == id)) {
    | exception Not_found => false
    | _contact => true
    };
  let getThreshold = (id, currContacts) =>
    switch (currContacts |> List.find(contact => fst(contact) == id)) {
    | exception Not_found => ""
    | testContact => snd(testContact)
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
                         checked={value |> isChecked(contact.contactId)}
                         onChange={e =>
                           onChange(
                             Form.TupleList(
                               updateTestContacts(
                                 ReactEvent.Form.target(e)##checked,
                                 (contact.contactId, "0"),
                                 value,
                               ),
                             ),
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
                           onChange(
                             Form.TupleList(
                               updateThreshold(
                                 (
                                   contact.contactId,
                                   ReactEvent.Form.target(e)##value,
                                 ),
                                 value,
                               ),
                             ),
                           )
                         }
                         value={value |> getThreshold(contact.contactId)}
                         disabled={!isChecked(contact.contactId, value)}
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
    {errorMsg != React.null
       ? <p className="text-red-500 text-xs italic"> errorMsg </p>
       : <p className="text-gray-600 text-xs italic">
           {"Who should be contacted upon error and after how many consecutive test failures"
            |> React.string}
         </p>}
  </div>;
};