[@react.component]
let make = () => {
  <nav className="flex items-center justify-between flex-wrap bg-teal-500 p-6">
    <div className="flex items-center flex-shrink-0 text-white mr-6">
      <span className="font-semibold text-xl tracking-tight">
        {"Pingr" |> React.string}
      </span>
    </div>
    <div className="w-full block flex-grow lg:flex lg:items-center lg:w-auto">
      <div className="text-sm lg:flex-grow">
        <a
          onClick={_ => Paths.goToDashboard()}
          className="cursor-pointer block mt-4 lg:inline-block lg:mt-0 text-teal-200 hover:text-white mr-4">
          {"Dashboard" |> React.string}
        </a>
        <a
          onClick={_ => Paths.goToTests()}
          className="cursor-pointer block mt-4 lg:inline-block lg:mt-0 text-teal-200 hover:text-white mr-4">
          {"Tests" |> React.string}
        </a>
        <a
          onClick={_ => Paths.goToContacts()}
          className="cursor-pointer block mt-4 lg:inline-block lg:mt-0 text-teal-200 hover:text-white mr-4">
          {"Contacts" |> React.string}
        </a>
      </div>
    </div>
  </nav>;
};