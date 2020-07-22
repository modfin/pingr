type width =
  | Full
  | Half;

[@react.component]
let make =
    (
      ~width: width,
      ~label,
      ~placeholder,
      ~options: list((string, string)),
      ~onChange,
      ~errorMsg,
      ~infoText,
      ~value,
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
    <div className="relative">
      <select
        value={
          switch (value) {
          | Form.Str(s) => s
          | _ => ""
          }
        }
        onChange={e =>
          Form.Str(ReactEvent.Form.target(e)##value) |> onChange
        }
        className="block appearance-none w-full bg-gray-200 border border-gray-400 text-gray-700 py-3 px-4 pr-8 rounded leading-tight focus:outline-none focus:bg-white focus:border-gray-500">
        <option value="" disabled=true hidden=true>
          {placeholder |> React.string}
        </option>
        {options
         |> List.map(((name, value)) =>
              <option key=value value> {name |> React.string} </option>
            )
         |> Array.of_list
         |> React.array}
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
    {errorMsg != React.null
       ? <p className="text-red-500 text-xs italic"> errorMsg </p>
       : <p className="text-gray-600 text-xs italic">
           {infoText |> React.string}
         </p>}
  </div>;
};