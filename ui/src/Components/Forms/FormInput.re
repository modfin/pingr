type inputTypes =
  | Text
  | Password
  | Number;

type width =
  | Full
  | Half;

[@react.component]
let make =
    (
      ~type_: inputTypes,
      ~width: width,
      ~onChange: Form.t => unit,
      ~value,
      ~label,
      ~errorMsg,
      ~infoText,
      ~placeholder=?,
    ) => {
  <div
    className={
      switch (width) {
      | Full => "w-full px-3 mb-6 md:mb-0"
      | Half => "w-full md:w-1/2 px-3 mb-6 md:mb-0"
      }
    }>
    <label
      className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
      {label |> React.string}
    </label>
    <input
      value={
        switch (type_) {
        | Password
        | Text =>
          switch (value) {
          | Form.Str(s) => s
          | _ => ""
          }
        | Number =>
          switch (value) {
          | Form.Int(i) => string_of_int(i)
          | _ => "0"
          }
        }
      }
      onChange={e =>
        (
          switch (type_) {
          | Password
          | Text => Form.Str(ReactEvent.Form.target(e)##value)
          | Number =>
            try(Form.Int(int_of_string(ReactEvent.Form.target(e)##value))) {
            | _ => Form.Int(0)
            }
          }
        )
        |> onChange
      }
      className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-400 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
      type_={
        switch (type_) {
        | Password => "password"
        | Text => "text"
        | Number => "number"
        }
      }
      placeholder={
        switch (placeholder) {
        | None => ""
        | Some(ph) => ph
        }
      }
    />
    {errorMsg != React.null
       ? <p className="text-red-500 text-xs italic"> errorMsg </p>
       : <p className="text-gray-600 text-xs italic">
           {infoText |> React.string}
         </p>}
  </div>;
};