use std::fs::ReadDir;
use std::path::Path;
use std::fs;
use std::io;
use std::vec;

const READDir: &str = "";

fn main() {
    // 各部のリストを作る
    let path = Path::new(READDir);
    let patterns = read_dir(path);
    println!("{:?}", patterns);

    for result in patterns.iter() {
        println!(r#"{:?}"#, result);
    }

    // // 各部ごとのファイルリストを作る
    // let mut filelist = Vec::new();
    // for result in &patterns {
    //     let results = read_files(result.join(READDir))?;
    //     println!("{:?}", results);
    // }

    println!("Hello, world!");
}

// imagesディレクトリ内の各部署ディレクトリの画像を走査する
fn read_dir<P: AsRef<Path>>(path: P) -> io::Result<Vec<String>> {
    Ok(fs::read_dir(path)?
        .filter_map(|entry| {
            let entry = entry.ok()?;
            if entry.file_type().ok()?.is_dir() {
                Some(entry.file_name().to_string_lossy().into_owned())
            } else {
                None
            }
        })
        .collect())
}
