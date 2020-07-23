[@react.component]
let make = (~label, ~placeholder, ~value, ~onChange, ~errorMsg, ~infoText) => {
  <div className="w-full px-3 mb-6 md:mb-0">
    <label
      className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
      {label |> React.string}
    </label>
    <textarea
      value={
        switch (value) {
        | Form.Str(s) => s
        | _ => ""
        }
      }
      onChange={e => Form.Str(ReactEvent.Form.target(e)##value) |> onChange}
      placeholder
      className="h-64 resize-y appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-400 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
    />
    {errorMsg != React.null
       ? <p className="text-red-500 text-xs italic"> errorMsg </p>
       : <p className="text-gray-600 text-xs italic">
           {infoText |> React.string}
         </p>}
  </div>;
};