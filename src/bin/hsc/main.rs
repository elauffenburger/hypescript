use std::{
    env::{self},
    fmt::Display,
    fs,
    io::{self, Read},
};

use hypescript::{emitter, parser::{self, Module}, util::rcref};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let src = {
        let mut buf = String::new();
        io::stdin().read_to_string(&mut buf)?;

        buf
    };

    let args = parse_args()?;
    let output_dir = args.output_dir.expect("output dir required!");

    let parser = parser::Parser::new(rcref(Module{}));
    let emitted = emitter::Emitter::new().emit(&[parser.parse(&src)?])?;

    for file in &emitted.files {
        write_emitted_file(&file, &output_dir)?;
    }

    Ok(())
}

fn write_emitted_file(
    file: &emitter::EmittedFile,
    dir: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    Ok(match file {
        emitter::EmittedFile::Dir { name, files } => {
            let dir = &format!("{dir}/{name}");
            fs::create_dir_all(dir)?;

            for file in files {
                write_emitted_file(&file, dir)?;
            }
        }
        emitter::EmittedFile::File { name, content } => {
            fs::create_dir_all(dir)?;

            fs::write(format!("{dir}/{name}"), content)?;
        }
    })
}

struct ParseArgsError {
    msg: String,
}

impl std::fmt::Debug for ParseArgsError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(&self.msg)
    }
}

impl Display for ParseArgsError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(&self.msg)
    }
}

impl std::error::Error for ParseArgsError {}

impl ParseArgsError {
    pub fn new(msg: &str) -> Box<ParseArgsError> {
        Box::new(ParseArgsError { msg: msg.into() })
    }
}

struct HscArgs {
    output_dir: Option<String>,
}

fn parse_args() -> Result<HscArgs, Box<dyn std::error::Error>> {
    let mut args = HscArgs { output_dir: None };

    let mut env_args = env::args().skip(1);
    while let Some(arg) = env_args.next() {
        match arg.as_str() {
            "-o" => match env_args.next() {
                Some(output_dir) => {
                    args.output_dir = Some(output_dir);
                }
                None => return Err(ParseArgsError::new("arg for -o required")),
            },
            _ => return Err(ParseArgsError::new(&format!("unknown arg {arg}"))),
        }
    }

    Ok(args)
}
