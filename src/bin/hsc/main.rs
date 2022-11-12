use std::{env, fmt::Display, fs};

use hypescript::{emitter, parser};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = parse_args()?;

    let output_dir = args.output_dir.expect("output dir required!");

    let parser = parser::Parser::new();
    let parsed_mods: Result<Vec<_>, parser::Error> = args
        .srcs
        .iter()
        .map(|path| {
            fs::read_to_string(path)
                .map_err(|err| Box::new(err) as parser::Error)
                .and_then(|src| parser.parse(&path, &src))
        })
        .collect();

    let emitted = emitter::Emitter::new().emit(&parsed_mods?)?;

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
    srcs: Vec<String>,
}

fn parse_args() -> Result<HscArgs, Box<dyn std::error::Error>> {
    let mut args = HscArgs {
        output_dir: None,
        srcs: vec![],
    };

    let mut env_args = env::args().skip(1);
    while let Some(arg) = env_args.next() {
        match arg.as_str() {
            "-o" => match env_args.next() {
                Some(output_dir) => {
                    args.output_dir = Some(output_dir);
                }
                None => return Err(ParseArgsError::new("arg for -o required")),
            },
            arg @ _ => {
                if arg.starts_with("-") {
                    return Err(ParseArgsError::new(&format!("unknown arg {arg}")));
                }

                args.srcs.push(arg.into())
            }
        }
    }

    Ok(args)
}
