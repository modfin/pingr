type testTypes =
  | HTTP
  | Prometheus
  | TLS
  | DNS
  | Ping
  | SSH
  | TCP
  | HTTPPush
  | PrometheusPush;

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

type contactState =
  | NotAsked
  | Loading
  | Success(list(Models.Contact.t))
  | Failure;

[@react.component]
let make = () => {
  let (responseMsg, setResponseMsg) = React.useState(_ => "");

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

  React.useEffect1(
    () => {
      switch (state) {
      | Loadable.Loading =>
        Api.fetchContactsWithCallback(result =>
          switch (result) {
          | None => dispatch(LoadFail("No contacts, add one above"))
          | Some(contacts) => dispatch(LoadSuccess(contacts))
          }
        )
      | _ => () /* Only fetch data when loading */
      };
      None;
    },
    [|state|],
  );
  <>
    <div className="relative bg-gray-400 my-4 p-1">
      <p className="text-xl font-bold"> {"Contacts" |> React.string} </p>
      <button
        onClick={_event => Paths.goToNewContact()}
        className="m-1 bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded">
        {"New contact" |> React.string}
      </button>
    </div>
    {switch (state) {
     | Loading => <div> {ReasonReact.string("Loading...")} </div>
     | Failed(msg) => <div> {ReasonReact.string(msg)} </div>
     | Success(contacts) =>
       <>
         <table className="table-auto text-left">
           <thead>
             <tr>
               <th className="px-4 py-2"> {"Name" |> React.string} </th>
               <th className="px-4 py-2"> {"Type" |> React.string} </th>
               <th className="px-4 py-2"> {"Url" |> React.string} </th>
             </tr>
           </thead>
           <tbody>
             Models.Contact.(
               {contacts
                |> List.map(contact => {
                     <tr key={contact.contactId}>
                       <td className="border px-4 py-2">
                         <a
                           className="no-underline text-blue-500 hover:underline cursor-pointer"
                           onClick={_event => Paths.goToTests()}>
                           {contact.contactName |> React.string}
                         </a>
                       </td>
                       <td className="border px-4 py-2">
                         {contact.contactType |> React.string}
                       </td>
                       <td className="border px-4 py-2">
                         {contact.contactUrl |> React.string}
                       </td>
                       <td>
                         <button
                           onClick={_e =>
                             Paths.goToEditContact(contact.contactId)
                           }
                           className="m-1 bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
                           {"Edit" |> React.string}
                         </button>
                       </td>
                       <td>
                         <button
                           onClick={_e =>
                             Api.deleteContact(contact.contactId, response => {
                               switch (response) {
                               | Success(msg) =>
                                 setResponseMsg(_ => msg);
                                 dispatch(LoadData);
                               | Error(msg) => setResponseMsg(_ => msg)
                               | SuccessJSON(_json) => () /* will not happen */
                               }
                             })
                           }
                           className="bg-red-500 hover:bg-red-700 text-white py-1 px-2 rounded">
                           {"Delete" |> React.string}
                         </button>
                       </td>
                     </tr>
                   })
                |> Array.of_list
                |> React.array}
             )
           </tbody>
         </table>
         {responseMsg != ""
            ? <p className="text-gray-500 text-xs italic">
                {responseMsg |> React.string}
              </p>
            : React.null}
       </>
     }}
  </>;
};